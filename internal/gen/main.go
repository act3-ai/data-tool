// Package main is a fake package for generating code.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/act3-ai/data-tool/pkg/apis"
	"github.com/act3-ai/go-common/pkg/genschema"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Must specify a target directory for schema generation.")
	}

	// Generate JSON Schema definitions
	if err := genschema.GenerateGroupSchemas(
		os.Args[1],
		apis.NewScheme(),
		[]string{"config.dt.act3-ace.io"},
		"github.com/act3-ai/data-tool",
	); err != nil {
		log.Fatal(fmt.Errorf("JSON Schema generation failed: %w", err))
	}
}
