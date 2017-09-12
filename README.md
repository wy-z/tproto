# tproto
[![Build Status](https://travis-ci.org/wy-z/tproto.svg?branch=master)](https://travis-ci.org/wy-z/tproto) [![GoDoc](https://godoc.org/github.com/wy-z/tproto?status.svg)](http://godoc.org/github.com/wy-z/tproto) [![Go Report Card](https://goreportcard.com/badge/github.com/wy-z/tproto)](https://goreportcard.com/report/github.com/wy-z/tproto)

Parse golang data structure into proto3.

## Installation
```
go get github.com/wy-z/tproto
```
Or
```
import "github.com/wy-z/tproto/tproto" # see cli/cli.go
```

## Usage
```
NAME:
   tproto - Parse golang data structure into proto3.

USAGE:
   tproto [global options] command [command options] [arguments...]

VERSION:
   0.0.1

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --package PKG, --pkg PKG        (required) package url PKG
   --expression EXPR, --expr EXPR  (required) type expression EXPR
   --proto-package PP, --pp PP     (required) proto package PP
   --proto-file PF, --pf PF        load messages from proto file PF
   --help, -h                      show help
   --version, -v                   print the version
```

## QuickStart

`tproto -pkg github.com/wy-z/tproto/samples -expr BasicTypes -pp samples`

## Samples

see `github.com/wy-z/tproto/samples/source`

## Test

```
go get -u github.com/jteeuwen/go-bindata/...
go generate ./samples && go test -v ./tproto
```