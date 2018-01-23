# tspec
[![Build Status](https://travis-ci.org/wy-z/tspec.svg?branch=master)](https://travis-ci.org/wy-z/tspec) [![GoDoc](https://godoc.org/github.com/wy-z/tspec?status.svg)](http://godoc.org/github.com/wy-z/tspec) [![Go Report Card](https://goreportcard.com/badge/github.com/wy-z/tspec)](https://goreportcard.com/report/github.com/wy-z/tspec)

Parse golang data structure into json schema.

## Installation
```
go get github.com/wy-z/tspec/...
```
Or
```
import "github.com/wy-z/tspec/tspec" # see cmd/tspec/cli.go
```

## Usage
```
NAME:
   TSpec - Parse golang data structure into json schema.

USAGE:
   tspec [global options] command [command options] [arguments...]

VERSION:
   2.2.0

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --package PKG, -p PKG                package path PKG (default: ".")
   --expressions EXPRS, --exprs EXPRS   (any-of required) type expressions, seperated by ',' EXPRS
   --decorator DECORATOR, -d DECORATOR  (any-of required) parse package with decorator DECORATOR
   --ref-prefix PREFIX, --rp PREFIX     the prefix of ref url PREFIX (default: "#/definitions/")
   --ignore-json-tag, --igt             ignore json tag
   --help, -h                           show help
   --version, -v                        print the version
```

## QuickStart

`tspec -p github.com/wy-z/tspec/samples -exprs BasicTypes,NormalStruct`
Or
`tspec -p github.com/wy-z/tspec/samples  BasicTypes NormalStruct`

## Samples

see `github.com/wy-z/tspec/samples/source`

## Test

```
go get -u github.com/jteeuwen/go-bindata/...
go generate ./samples && go test -v ./tspec
```
