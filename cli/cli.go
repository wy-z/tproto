package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli"
	"github.com/wy-z/tproto/tproto"
)

type cliOpts struct {
	PkgPath   string
	TypeExprs string
	ProtoPkg  string
	ProtoFile string
	JSONTag   bool
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
			Usage:       "(required) type expressions, seperated by ',', `EXPRS`",
			Destination: &opts.TypeExprs,
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
		if opts.TypeExprs == "" || opts.ProtoPkg == "" {
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

		for _, expr := range strings.Split(opts.TypeExprs, ",") {
			expr = strings.TrimSpace(expr)
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
