package tproto_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/wy-z/tproto/samples"
	"github.com/wy-z/tproto/tproto"
)

func TestTProto(t *testing.T) {
	suite.Run(t, new(TProtoTestSuite))
}

type TProtoTestSuite struct {
	suite.Suite
	parser *tproto.Parser
	pkg    string
}

func (s *TProtoTestSuite) SetupTest() {
	s.parser = tproto.NewParser()
	s.pkg = "github.com/wy-z/tproto/samples"
}

const samplesProtoPkg = "samples"

func (s *TProtoTestSuite) testParse(typeStr, assert string) {
	require := s.Require()

	schema, err := s.parser.Parse(s.pkg, typeStr)
	require.NoError(err)
	require.NotNil(schema)

	buf := s.parser.RenderProto(samplesProtoPkg)
	require.NoError(err)
	require.Equal(string(bytes.TrimSpace(samples.MustAsset(assert))),
		string(bytes.TrimSpace(buf.Bytes())))
	s.parser.Reset()
}

func (s *TProtoTestSuite) TestParse() {
	s.testParse("BasicTypes", "source/basic_types.proto")
	s.testParse("NormalStruct", "source/normal_struct.proto")
	s.testParse("StructWithNoExportField", "source/struct_with_no_export_field.proto")
	s.testParse("StructWithAnonymousField", "source/struct_with_anonymous_field.proto")
	s.testParse("StructWithCircularReference", "source/struct_with_circular_reference.proto")
	s.testParse("StructWithInheritance", "source/struct_with_inheritance.proto")
}
