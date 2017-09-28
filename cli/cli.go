package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli"
	"github.com/wy-z/tproto/tproto"
	"github.com/wy-z/tspec/tspec"
)

type cliOpts struct {
	PkgPath   string
	TypeExprs string
	ProtoPkg  string
	ProtoFile string
	JSONTag   bool
	Decorator string
}

//Run runs tproto
func Run(version string) {
	app := cli.NewApp()
	app.Name = "tproto"
	app.Version = version
	app.Usage = "Parse golang data structure into proto3."

	opts := new(cliOpts)
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "package, p",
			Usage:       "package path `PKG`",
			Value:       ".",
			Destination: &opts.PkgPath,
		},
		cli.StringFlag{
			Name:        "expressions, exprs",
			Usage:       "(any-of required) type expressions, seperated by ',' `EXPRS`",
			Destination: &opts.TypeExprs,
		},
		cli.StringFlag{
			Name:        "decorator, d",
			Usage:       "(any-of required) parse package with decorator `DECORATOR`",
			Destination: &opts.Decorator,
		},
		cli.StringFlag{
			Name:        "proto-package, pp",
			Usage:       "(required) proto package `PP`",
			Destination: &opts.ProtoPkg,
		},
		cli.StringFlag{
			Name:        "proto-file, pf",
			Usage:       "load messages from proto file `PF`",
			Destination: &opts.ProtoFile,
		},
		cli.BoolFlag{
			Name:        "json-tag, jt",
			Usage:       "don't ignore json tag",
			Destination: &opts.JSONTag,
		},
	}
	app.Action = func(c *cli.Context) (err error) {
		if c.NArg() > 0 {
			opts.TypeExprs = strings.Join(c.Args(), ",")
		}
		if opts.ProtoPkg == "" || (opts.TypeExprs == "" && opts.Decorator == "") {
			cli.ShowAppHelp(c)
			return
		}

		parser := tproto.NewParser()
		parserOpts := tproto.DefaultParserOptions
		parserOpts.IgnoreJSONTag = !opts.JSONTag
		parser.Options(parserOpts)

		if opts.ProtoFile != "" {
			err = parser.LoadProtoFile(opts.ProtoFile)
			if err != nil {
				msg := fmt.Sprintf("failed to load proto file %s: %s", opts.ProtoFile, err)
				err = cli.NewExitError(msg, 1)
				return
			}
		}

		exprs := make([]string, 0, 2)
		for _, expr := range strings.Split(opts.TypeExprs, ",") {
			expr = strings.TrimSpace(expr)
			if expr == "" {
				continue
			}
			exprs = append(exprs, expr)
		}
		if opts.Decorator != "" {
			pkg, e := tspec.NewParser().Import(opts.PkgPath)
			if e != nil {
				msg := fmt.Sprintf("failed to import pkg '%s': %s", opts.PkgPath, e)
				err = cli.NewExitError(msg, 1)
				return
			}
			objs, e := tspec.ParsePkgWithDecorator(pkg, opts.Decorator)
			if e != nil {
				msg := fmt.Sprintf("failed to parse pkg with decorator, %s", e)
				err = cli.NewExitError(msg, 1)
				return
			}
			for k := range objs {
				exprs = append(exprs, k)
			}
		}

		for _, expr := range exprs {
			_, err = parser.Parse(opts.PkgPath, expr)
			if err != nil {
				msg := fmt.Sprintf("failed to parse type expr %s: %s", expr, err)
				err = cli.NewExitError(msg, 1)
				return
			}
		}

		fmt.Println(parser.RenderProto(opts.ProtoPkg).String())
		return
	}

	app.Run(os.Args)
}
