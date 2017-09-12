package cli

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
	"github.com/wy-z/tproto/tproto"
)

type cliOpts struct {
	PkgURL    string
	TypeExpr  string
	ProtoPkg  string
	ProtoFile string
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
			Name:        "package, pkg",
			Usage:       "(required) package url `PKG`",
			Destination: &opts.PkgURL,
		},
		cli.StringFlag{
			Name:        "expression, expr",
			Usage:       "(required) type expression `EXPR`",
			Destination: &opts.TypeExpr,
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
	}
	app.Action = func(c *cli.Context) (err error) {
		if opts.PkgURL == "" || opts.TypeExpr == "" || opts.ProtoPkg == "" {
			cli.ShowAppHelp(c)
			return
		}

		parser := tproto.NewParser()
		if opts.ProtoFile != "" {
			err = parser.LoadProtoFile(opts.ProtoFile)
			if err != nil {
				msg := fmt.Sprintf("failed to load proto file %s: %s", opts.ProtoFile, err)
				err = cli.NewExitError(msg, 1)
				return
			}
		}
		_, err = parser.Parse(opts.PkgURL, opts.TypeExpr)
		if err != nil {
			msg := fmt.Sprintf("failed to parse type expr %s: %+v", opts.TypeExpr, err)
			err = cli.NewExitError(msg, 1)
			return
		}
		fmt.Println(parser.RenderProto(opts.ProtoPkg).String())
		return
	}

	app.Run(os.Args)
}
