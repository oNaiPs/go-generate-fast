//go:build tools

package main

import (
	_ "gotest.tools/gotestsum"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/cheekybits/genny"
	_ "go.uber.org/mock/mockgen"
	_ "github.com/go-bindata/go-bindata"
	_ "github.com/mjibson/esc"
	_ "github.com/99designs/gqlgen"
	_ "golang.org/x/tools/cmd/stringer"
)