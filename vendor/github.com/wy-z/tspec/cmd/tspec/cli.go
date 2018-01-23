package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli"
	"github.com/wy-z/tspec/tspec"
)

type cliOpts struct {
	PkgPath       string
	TypeExprs     string
	RefPrefix     string
	IgnoreJSONTag bool
	Decorator     string
}

//Run runs tspec
func Run(version string) {
	app := cli.NewApp()
	app.Name = "TSpec"
	app.Version = version
	app.Usage = "Parse golang data structure into json schema."

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
			Name:        "ref-prefix, rp",
			Usage:       "the prefix of ref url `PREFIX`",
			Value:       tspec.DefaultRefPrefix,
			Destination: &opts.RefPrefix,
		},
		cli.BoolFlag{
			Name:        "ignore-json-tag, igt",
			Usage:       "ignore json tag",
			Destination: &opts.IgnoreJSONTag,
		},
	}
	app.Action = func(c *cli.Context) (err error) {
		if c.NArg() > 0 {
			opts.TypeExprs = strings.Join(c.Args(), ",")
		}
		if opts.TypeExprs == "" && opts.Decorator == "" {
			cli.ShowAppHelp(c)
			return
		}

		parser := tspec.NewParser()
		parserOpts := tspec.DefaultParserOptions
		if opts.RefPrefix != "" {
			parserOpts.RefPrefix = opts.RefPrefix
		}
		parserOpts.IgnoreJSONTag = opts.IgnoreJSONTag
		parser.Options(parserOpts)

		pkg, err := parser.Import(opts.PkgPath)
		if err != nil {
			msg := fmt.Sprintf("failed to import pkg '%s': %s", opts.PkgPath, err)
			err = cli.NewExitError(msg, 1)
			return
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
			_, err = parser.Parse(pkg, expr)
			if err != nil {
				msg := fmt.Sprintf("failed to parse type expr %s: %s", opts.TypeExprs, err)
				err = cli.NewExitError(msg, 1)
				return
			}
		}

		defs := parser.Definitions()
		bytes, err := json.MarshalIndent(defs, "", "\t")
		if err != nil {
			msg := fmt.Sprintf("failed to marshal definitions: %s", err)
			err = cli.NewExitError(msg, 1)
			return
		}
		fmt.Println(string(bytes))
		return
	}

	app.Run(os.Args)
}
