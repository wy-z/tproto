package tproto

import (
	"bytes"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/emicklei/proto"
	"github.com/go-openapi/spec"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/wy-z/tspec/tspec"
)

var jsonProtoTypeMap = map[string]string{
	"integer":          "int64",
	"integer:int32":    "int32",
	"integer:int64":    "int64",
	"number":           "double",
	"number:float":     "float",
	"number:double":    "double",
	"string":           "string",
	"string:byte":      "bytes",
	"string:binary":    "bytes",
	"boolean":          "bool",
	"string:date":      "string",
	"string:date-time": "string",
}

// ParserOptions defines tproto parser options
type ParserOptions struct {
	tspec.ParserOptions
}

const tspecRefPrefix = "#/"

// DefaultParserOptions defines default tproto parser options
var DefaultParserOptions = ParserOptions{
	ParserOptions: tspec.ParserOptions{
		IgnoreJSONTag: false,
		RefPrefix:     tspecRefPrefix,
	},
}

// Parser defines tproto parser
type Parser struct {
	messages map[string]*proto.Message
	opts     ParserOptions
	lock     sync.Mutex
}

// NewParser returns inited tproto parser
func NewParser() (parser *Parser) {
	parser = new(Parser)
	parser.messages = make(map[string]*proto.Message)
	parser.opts = DefaultParserOptions
	return
}

// Options gets or sets parser options
func (t *Parser) Options(opts ...ParserOptions) ParserOptions {
	if len(opts) != 0 {
		t.opts = opts[0]
	}
	return t.opts
}

// Messages returns all messages
func (t *Parser) Messages() map[string]*proto.Message {
	return t.messages
}

// SetMessages set messages
func (t *Parser) SetMessages(messages map[string]*proto.Message) {
	t.messages = messages
	return
}

// LoadProtoFile loads messages from proto file
func (t *Parser) LoadProtoFile(path string) (err error) {
	p, err := ParseProtoFile(path)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	for _, each := range p.Elements {
		if msg, ok := each.(*proto.Message); ok {
			t.messages[msg.Name] = msg
		}
	}
	return
}

// Reset cleans all messages
func (t *Parser) Reset() {
	t.messages = make(map[string]*proto.Message)
	return
}

// ProtoSyntax defines proto synatax
const ProtoSyntax = "proto3"

// RenderProto renders proto messages
func (t *Parser) RenderProto(protoPkg string) (buf *bytes.Buffer) {
	p := new(proto.Proto)
	p.Elements = append(p.Elements, &proto.Syntax{
		Value: ProtoSyntax,
	})
	p.Elements = append(p.Elements, &proto.Package{
		Name: protoPkg,
	})

	keys := make(sort.StringSlice, 0, 2)
	for k := range t.messages {
		keys = append(keys, k)
	}
	keys.Sort()
	for _, k := range keys {
		p.Elements = append(p.Elements, t.messages[k])
	}

	buf = bytes.NewBuffer(nil)
	proto.NewFormatter(buf, "  ").Format(p)
	return
}

func (t *Parser) parseDefinitionField(title string, field *spec.Schema, sequence int) (fieldProto proto.Visitee, err error) {
	typeStr := schemaTypeStr(field)
	var isMap, isArray, isRef bool
	if typeStr == "object" && field.AdditionalProperties != nil {
		field = field.AdditionalProperties.Schema
		typeStr = schemaTypeStr(field)
		isMap = true
	} else if typeStr == "array" {
		field = field.Items.Schema
		typeStr = schemaTypeStr(field)
		isArray = true
	}
	if field.Ref.String() != "" {
		isRef = true
	}

	f := new(proto.Field)
	f.Name = title
	f.Sequence = sequence
	if typeStr == emptyType {
		log.Warnf("ignored unsupported type %s", title)
		return
	}

	if protoType, ok := jsonProtoTypeMap[typeStr]; ok {
		f.Type = protoType
	} else if isRef {
		f.Type = typeStr
	} else {
		err = errors.Errorf("unsupported type %s", typeStr)
		return
	}

	if isMap {
		fp := &proto.MapField{
			Field:   f,
			KeyType: "string",
		}
		fieldProto = fp
	} else {
		fp := &proto.NormalField{
			Field:    f,
			Repeated: isArray,
		}
		fieldProto = fp
	}
	return
}

func (t *Parser) parseDefinition(def *spec.Schema) (message *proto.Message, err error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if m, ok := t.messages[def.Title]; ok {
		message = m
		return
	}

	message = new(proto.Message)
	message.Name = def.Title
	typeStr := schemaTypeStr(def)
	switch typeStr {
	case "object":
		fields := schemaAllProperties(def)
		keys := make(sort.StringSlice, 0, 12)
		for k := range fields {
			keys = append(keys, k)
		}
		keys.Sort()
		for i, k := range keys {
			f, e := t.parseDefinitionField(k, fields[k], i+1)
			if e != nil {
				err = errors.WithStack(e)
				return
			}
			if f != nil {
				message.Elements = append(message.Elements, f)
			}
		}
	default:
		err = errors.Errorf("unsupported type %s", typeStr)
		return
	}
	t.messages[def.Title] = message
	return
}

// Parse parses golang type expr
func (t *Parser) Parse(pkgPath, typeExpr string) (message *proto.Message, err error) {
	def, defs, err := t.parseTypeExpr(pkgPath, typeExpr)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	for _, d := range defs {
		delete(t.messages, d.Title)
		_, e := t.parseDefinition(&d)
		if e != nil {
			err = errors.WithStack(e)
			return
		}
	}
	message = t.messages[def.Title]
	return
}

func (t *Parser) parseTypeExpr(pkgPath, typeExpr string) (def *spec.Schema, defs spec.Definitions, err error) {
	parser := tspec.NewParser()
	parser.Options(t.opts.ParserOptions)

	pkg, err := parser.Import(pkgPath)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	def, err = parser.Parse(pkg, typeExpr)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	defs = parser.Definitions()
	return
}

// ParseProtoFile parses proto file
func ParseProtoFile(path string) (p *proto.Proto, err error) {
	reader, err := os.Open(path)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	protoParser := proto.NewParser(reader)
	p, err = protoParser.Parse()
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

const (
	emptyType = "empty"
)

func schemaTypeStr(schema *spec.Schema) string {
	if len(schema.Type) == 0 && schema.Ref.String() == "" {
		return emptyType
	}

	l := make([]string, 0, 2)
	if len(schema.Type) > 0 {
		l = append(l, schema.Type[0])
	} else {
		l = append(l, schema.Ref.String()[len(tspecRefPrefix):])
	}
	if schema.Format != "" {
		l = append(l, schema.Format)
	}
	return strings.Join(l, ":")
}

func schemaAllProperties(schema *spec.Schema) (props map[string]*spec.Schema) {
	props = make(map[string]*spec.Schema)
	for _, s := range schema.AllOf {
		for k, v := range schemaAllProperties(&s) {
			props[k] = v
		}
	}
	for k, v := range schema.Properties {
		vc := v
		props[k] = &vc
	}
	return
}
