//go:build tools
// +build tools

package tools

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/ory/go-acc/cmd"
	_ "golang.org/x/tools/cmd/goimports"
)
