// Package gen is a fake package for generating code.
package gen

//go:generate rm -rf docs/apis/schemas
//go:generate go run git.act3-ace.com/ace/data/schema/cmd/genschema docs/apis/schemas
//go:generate go run internal/gen/main.go docs/apis/schemas
//go:generate tool/controller-gen object paths=./...
