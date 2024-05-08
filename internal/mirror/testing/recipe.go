// Package testing helps generate test data for use in testing telemetry
package testing

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/schema"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/errdef"

	"git.act3-ace.com/ace/data/tool/internal/orasutil"
	"git.act3-ace.com/ace/data/tool/internal/ui"
)

// templateFile executes the template for the given template file.
func templateFile(t *template.Template, out string) error {
	raw, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("unable to write templated output file: %w", err)
	}
	defer raw.Close()

	values := struct {
		Filename string
	}{
		Filename: filepath.Base(out),
	}

	// execute the template
	if err := t.Execute(raw, values); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	return raw.Close()
}

// validateFile the results if they start with index, manifest, or config.
func validateFile(fsys fs.FS, filename, mediaType string) error {
	// base := filepath.Base(filename)

	val := schema.Validator(mediaType)

	if len(val) == 0 {
		// unable to validate this content
		return nil
	}

	raw, err := fsys.Open(filename)
	if err != nil {
		return fmt.Errorf("validating file: %w", err)
	}
	defer raw.Close()

	if err := val.Validate(raw); err != nil {
		// HACK  there seems to be no identifiable error that we can use errors.Is with ofr media types that are not able to be validated.
		if !strings.Contains(err.Error(), "unexpectedly not available in fsLoaderFactory") {
			return fmt.Errorf("validating %q: %w", filename, err)
		}
	}

	// TODO check that the "mediaType" field is the same as the mediaType passed in (for manifests only)
	return raw.Close()
}

type recipeStep struct {
	File      string           `json:"file"`
	MediaType string           `json:"mediaType"`
	Algorithm digest.Algorithm `json:"algorithm"`
	Tag       string           `json:"tag"`
}

// ProcessRecipe processes the steps in the recipe ad recipePath outputting the contents to ociDir.
// Optionally validating generated data.
func ProcessRecipe(ctx context.Context, recipePath, ociDir string, validate bool) error {
	rootUI := ui.FromContextOrNoop(ctx)

	recipeFile, err := os.Open(recipePath)
	if err != nil {
		return fmt.Errorf("opening recipe file: %w", err)
	}
	defer recipeFile.Close()

	// parse the file
	decoder := json.NewDecoder(recipeFile)

	dir := filepath.Dir(recipePath)
	fsh := &filesystemHelper{
		fsys: os.DirFS(dir),
	}

	// TemplateFuncs are used during templating the .tmpl files
	var templateFuncs = template.FuncMap{
		"FileDescriptor": fsh.FileDescriptor,
		"FileDigest":     fsh.FileDigest,
		"Tar":            fsh.Tar,
		"ToData":         func(input string) []byte { return []byte(input) },
		"Gzip":           gzipHelper,
		"Zstd":           zstdHelper,
	}

	t, err := template.New("root").
		Funcs(sprig.HermeticTxtFuncMap()).
		Funcs(templateFuncs).ParseFS(fsh.fsys, "*.tmpl")
	if err != nil {
		return fmt.Errorf("parsing templates in %s: %w", dir, err)
	}

	ociStore, err := oci.NewWithContext(ctx, ociDir)
	if err != nil {
		return fmt.Errorf("creating OCI layout at %q: %w", ociDir, err)
	}

	// double check that we have all the successors in the CAS.
	// We want to fail as early as possible.
	store := &orasutil.CheckedStorage{Target: ociStore}

	i := 0
	for {
		i++
		// Read a JSON line
		step := &recipeStep{}
		err := decoder.Decode(&step)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		if err := processStep(ctx, t, fsh, store.Target, validate, dir, step); err != nil {
			return fmt.Errorf("step %d: %w", i, err)
		}
	}

	rootUI.Info("OCI data available at: ", ociDir)
	return recipeFile.Close()
}

func processStep(ctx context.Context, t *template.Template, fsh *filesystemHelper, store oras.Target, validate bool, dir string, step *recipeStep) error {
	rootUI := ui.FromContextOrNoop(ctx)

	task := rootUI.SubTask(step.File)
	defer task.Complete()

	task.Info("Processing")

	if _, ok := fsh.seen[step.File]; ok {
		// This file has already been depended upon.
		return fmt.Errorf("file %s has already been depended upon by a file already processed, the recipe must not be in topological order", step.File)
	}

	if tt := t.Lookup(step.File + ".tmpl"); tt != nil {
		task.Info("Templating")
		if err := templateFile(tt, filepath.Join(dir, step.File)); err != nil {
			return err
		}
	}

	if validate {
		if err := validateFile(fsh.fsys, step.File, step.MediaType); err != nil {
			return err
		}
	}

	data, err := fs.ReadFile(fsh.fsys, step.File)
	if err != nil {
		return fmt.Errorf("reading file content: %w", err)
	}

	// compute the digest of the output file
	desc, err := fsh.fileDescriptorFromData(data, step.MediaType, step.Algorithm)
	if err != nil {
		return err
	}

	task.Info("Digest ", desc.Digest)

	// Push the file to OCI
	if err := store.Push(ctx, *desc, bytes.NewReader(data)); err != nil {
		// Push() fails if the blob is already present (that is not the correct behavior)
		if !errors.Is(err, errdef.ErrAlreadyExists) {
			return fmt.Errorf("pushing file: %w", err)
		}
	}
	task.Info("Pushed ", desc.MediaType)

	if step.Tag != "" {
		// tag as well
		if err := store.Tag(ctx, *desc, step.Tag); err != nil {
			return fmt.Errorf("tagging file: %w", err)
		}
		task.Info("Created tag ", step.Tag)
	}

	return nil
}
