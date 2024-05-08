// Package main is a fake package for generating code.
package main

import (
	"fmt"
	"log"
	"os"

	"git.act3-ace.com/ace/data/tool/pkg/apis"
	"git.act3-ace.com/ace/go-common/pkg/genschema"
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
		"git.act3-ace.com/ace/data/tool",
	); err != nil {
		log.Fatal(fmt.Errorf("JSON Schema generation failed: %w", err))
	}
}
