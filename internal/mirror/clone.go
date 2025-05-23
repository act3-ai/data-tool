// Package mirror implements the logic for the mirror commands.
package mirror

import (
	"context"
	"fmt"
	"log/slog"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sourcegraph/conc/pool"
	"oras.land/oras-go/v2"

	"github.com/act3-ai/data-tool/internal/print"
	"github.com/act3-ai/data-tool/internal/ref"
	dtreg "github.com/act3-ai/data-tool/internal/registry"
	"github.com/act3-ai/data-tool/internal/ui"
	reg "github.com/act3-ai/data-tool/pkg/registry"
)

// CloneOptions define the options required to run a Clone operation.
type CloneOptions struct {
	MappingSpec     string
	Selectors       []string
	ConcurrentHTTP  int
	Platforms       []string
	Log             *slog.Logger
	SourceFile      string
	RootUI          *ui.Task
	Targeter        reg.GraphTargeter
	Recursive       bool
	DryRun          bool
	ContinueOnError bool
}

// Clone will take a list of OCI references and scatter them according to the mapping spec.
func Clone(ctx context.Context, opts CloneOptions) error { //nolint:gocognit
	mapper, err := newMapper(opts.MappingSpec)
	if err != nil {
		return fmt.Errorf("creating the mapper: %w", err)
	}

	filters, err := parseFilters(opts.Selectors)
	if err != nil {
		return err
	}

	// throw the platforms in a map for easy querying
	var platforms []*ocispec.Platform
	if len(opts.Platforms) != 0 {
		platforms, err = parsePlatforms(opts.Platforms)
		if err != nil {
			return fmt.Errorf("parsing the platforms: %w", err)
		}
	}

	opts.Log.InfoContext(ctx, "Opening repository source file", "path", opts.SourceFile)
	sourceList, err := ProcessSourcesFile(ctx, opts.SourceFile, filters, opts.ConcurrentHTTP)
	if err != nil {
		return err
	}

	var p *pool.ContextPool
	if opts.ContinueOnError {
		p = pool.New().WithErrors().WithContext(ctx)
	} else {
		p = pool.New().WithErrors().WithContext(ctx).WithCancelOnError().WithFirstError()
	}
	p = p.WithMaxGoroutines(max(opts.ConcurrentHTTP/2, 1))

	wt := &WorkTracker{}
	var i int
	for _, src := range sourceList {
		i++
		task := opts.RootUI.SubTask(fmt.Sprintf("Source %d", i))
		p.Go(func(ctx context.Context) error {
			defer task.Complete()

			srcTarget, err := opts.Targeter.GraphTarget(ctx, src.Name)
			if err != nil {
				return err
			}

			// resolve the endpoint if necessary
			srcRef, err := dtreg.ParseEndpointOrDefault(opts.Targeter, src.Name)
			if err != nil {
				return err
			}

			// we fetch the reference in case it is a multi-architecture index
			// ensure we pass the full reference in the case srcTarget is an endpointResolver
			desc, err := srcTarget.Resolve(ctx, src.Name)
			if err != nil {
				return fmt.Errorf("error resolving the source: %w", err)
			}

			desc, err = annotateManifest(src.Name, desc, src.Labels, nil)
			if err != nil {
				return err
			}

			destinations, err := mapper(desc)
			if err != nil {
				return err
			}

			if len(destinations) == 0 {
				return nil
			}

			task.Infof("Copying %s", src.Name)
			var destCount int
			for _, destName := range destinations {
				destCount++

				destTarget, err := opts.Targeter.GraphTarget(ctx, destName)
				if err != nil {
					return fmt.Errorf("initializing destination graph target: %w", err)
				}

				// resolve the endpoint if necessary
				destRef, err := dtreg.ParseEndpointOrDefault(opts.Targeter, destName)
				if err != nil {
					return err
				}

				copyOpts := oras.CopyGraphOptions{
					MountFrom: mountFrom(srcRef, destRef),
					OnMounted: onMounted(opts.Log),
				}
				c, err := NewCopier(ctx, opts.Log, srcTarget, destTarget, desc, opts.Recursive, platforms, copyOpts)
				if err != nil {
					return err
				}
				c.options.PostCopy = func(ctx context.Context, desc ocispec.Descriptor) error {
					wt.Add(desc)
					return nil
				}

				// destination registry might be the same in each case in which case reusing the same client would be beneficial, automatically set by cache
				destTask := task.SubTask(fmt.Sprintf("destination %d/%d", destCount, len(destinations)))
				destTask.Infof("sending %s to %s", desc.Annotations[ref.AnnotationSrcRef], destRef)
				if opts.DryRun {
					destTask.Complete()
					return nil
				}
				if platforms == nil {
					if err := Copy(ctx, c); err != nil {
						destTask.Complete()
						return err
					}
					tag := destRef.ReferenceOrDefault()
					// Tag will work if `tag` is an actual tag or a digest
					if err := destTarget.Tag(ctx, desc, tag); err != nil {
						destTask.Complete()
						return fmt.Errorf("tagging scattered image as %s: %w", tag, err)
					}
				} else {
					platformDescriptors, err := CopyFilterOnPlatform(ctx, c)
					if err != nil {
						destTask.Complete()
						return err
					}
					for _, d := range platformDescriptors {
						tag := destRef.ReferenceOrDefault()
						// Tag will work if `tag` is an actual tag or a digest
						if err := destTarget.Tag(ctx, d, tag); err != nil {
							destTask.Complete()
							return fmt.Errorf("tagging scattered image as %s: %w", tag, err)
						}
					}

				}
				destTask.Complete()
			}
			// }
			return nil
		})
	}

	if err := p.Wait(); err != nil {
		if opts.ContinueOnError {
			opts.RootUI.Info(err)
		} else {
			return err
		}
	}

	opts.RootUI.Infof("%s pushed for %d blobs", print.Bytes(wt.transferred.Load()), wt.blobs.Load())
	return nil
}
