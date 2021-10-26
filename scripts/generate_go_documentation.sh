#!/usr/bin/env bash

set -x

go install golang.org/x/tools/cmd/godoc@latest
go install gitlab.com/tslocum/godoc-static@latest
export GOPATH=$(go env GOPATH)
export PATH=$PATH:$GOPATH/bin
rm -rf godoc
mkdir godoc
godoc-static -destination=godoc .
