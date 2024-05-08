package mirror

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/sync/errgroup"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"

	"git.act3-ace.com/ace/data/tool/internal/mirror/encoding"
	"git.act3-ace.com/ace/data/tool/internal/print"
	"git.act3-ace.com/ace/data/tool/internal/ref"
	"git.act3-ace.com/ace/data/tool/internal/ui"
	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// GatherOptions specify the requirements to run a mirror gather operation.
type GatherOptions struct {
	Platforms      []string
	ConcurrentHTTP int
	DestTarget     oras.Target
	Log            *slog.Logger
	RootUI         *ui.Task
	SourceFile     string
	Dest           string
	Annotations    map[string]string
	IndexFallback  bool
	DestReference  registry.Reference
	Recursive      bool
	RepoFunc       func(context.Context, string) (*remote.Repository, error)
}

// Gather will take the references defined in a SourceFile and consolidate them to a destination target.
func Gather(ctx context.Context, dataToolVersion string, opts GatherOptions) error { //nolint:gocognit
	// throw the platforms in a map for easy querying
	var platforms []*ocispec.Platform
	if len(opts.Platforms) != 0 {
		plat, err := parsePlatforms(opts.Platforms)
		if err != nil {
			return fmt.Errorf("error parsing the platforms: %w", err)
		}
		platforms = append(platforms, plat...)
	}

	bt := &ByteTracker{}
	wt := &WorkTracker{}

	var manifestsMutex sync.Mutex
	var manifests []ocispec.Descriptor

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(opts.ConcurrentHTTP)

	opts.Log.InfoContext(ctx, "Opening repository source file", "path", opts.SourceFile)
	sourceList, err := processSourcesFile(ctx, opts.SourceFile, nil, opts.ConcurrentHTTP)
	if err != nil {
		return err
	}
	var i int
	for _, src := range sourceList {
		i++
		task := opts.RootUI.SubTask(fmt.Sprintf("Source %d", i))
		g.Go(func() error {
			defer task.Complete()
			task.Infof("Copying %s", src.name)
			log := logger.V(opts.Log, 1)
			log.InfoContext(gctx, "copying", "srcReference", "dest reference", src.name, opts.Dest)
			// TODO: add progress back in... use GraphCopyOptions pre and post manifest
			numBytes := atomic.Int64{}
			c, err := NewCopier(ctx, log, src.name, opts.Dest, nil, registry.Reference{}, ocispec.Descriptor{}, opts.DestTarget, opts.DestReference, opts.Recursive, platforms, opts.RepoFunc)
			if err != nil {
				return err
			}

			var descriptors []ocispec.Descriptor
			c.options.PreCopy = func(ctx context.Context, desc ocispec.Descriptor) error {
				// does the descriptor already exist?
				exists, err := c.dest.Exists(gctx, desc)
				if err != nil {
					return fmt.Errorf("checking existence of %s: %w", c.root.Digest.String(), err)
				}
				if exists {
					return oras.SkipNode
				}
				return nil
			}
			// record the bytes and number of blobs that were actually copied.
			c.options.PostCopy = func(ctx context.Context, desc ocispec.Descriptor) error {
				// copied bytes and blobs
				wt.Add(desc)
				return nil
			}

			if platforms == nil {
				if err := Copy(gctx, c); err != nil {
					return fmt.Errorf("copying from %s: %w", src.name, err)
				}
				desc, err := annotateManifest(src.name, c.root, src.labels, nil)
				if err != nil {
					return err
				}
				// count bytes
				if err := extractBlobs(ctx, bt.AddDescriptor, opts.DestTarget, desc); err != nil {
					return fmt.Errorf("counting bytes: %w", err)
				}
				descriptors = append(descriptors, desc)
			} else {
				platformDescriptors, err := CopyFilterOnPlatform(gctx, c)
				if err != nil {
					return fmt.Errorf("copying with specific platforms for source %s: %w", src.name, err)
				}
				for _, d := range platformDescriptors {
					d, err := annotateManifest(src.name, d, src.labels, c.originatingIndex)
					if err != nil {
						return err
					}
					// count bytes
					if err := extractBlobs(ctx, bt.AddDescriptor, opts.DestTarget, d); err != nil {
						return fmt.Errorf("counting bytes: %w", err)
					}
					descriptors = append(descriptors, d)
				}
			}

			task.Infof("Copied %s", print.Bytes(numBytes.Load()))

			manifestsMutex.Lock()
			manifests = append(manifests, descriptors...)
			manifestsMutex.Unlock()

			return nil
		})
	}

	if err = g.Wait(); err != nil {
		return err
	}

	// set the ace-dt version, size, and deduplicated annotations
	opts.Annotations[encoding.AnnotationGatherVersion] = dataToolVersion
	opts.Annotations[encoding.AnnotationLayerSizeTotal] = fmt.Sprint(bt.Total)
	opts.Annotations[encoding.AnnotationLayerSizeDeduplicated] = fmt.Sprint(bt.Deduplicated)

	index := ocispec.Index{
		Versioned: specs.Versioned{
			SchemaVersion: 2,
		},
		MediaType:    ocispec.MediaTypeImageIndex,
		ArtifactType: encoding.MediaTypeGather,
		Manifests:    manifests,
		// add the user-defined annotations to the index annotations
		Annotations: opts.Annotations,
	}

	if opts.IndexFallback {
		encoding.IndexFallback(&index)
	}

	// marshal the index bytes for push
	b, err := json.Marshal(index)
	if err != nil {
		return fmt.Errorf("marshalling the index: %w", err)
	}

	indexDesc, err := oras.TagBytes(ctx, opts.DestTarget, ocispec.MediaTypeImageIndex, b, opts.DestReference.ReferenceOrDefault())
	if err != nil {
		return fmt.Errorf("pushing top-level index: %w", err)
	}
	// not deferring this log b/c it shouldn't display if gather fails
	opts.RootUI.Infof("Gathered %s (representing %s) to location: %s\n%s", print.Bytes(bt.Deduplicated), print.Bytes(bt.Total), opts.Dest, indexDesc.Digest)
	opts.RootUI.Infof("%s pushed for %d blobs", print.Bytes(wt.transferred.Load()), wt.blobs.Load())
	return nil
}

func annotateManifest(srcRef string, desc ocispec.Descriptor, labels map[string]string, sourceIndex []byte) (ocispec.Descriptor, error) {

	desc.Annotations = map[string]string{
		ref.AnnotationSrcRef: srcRef,
	}

	if len(labels) != 0 {
		// encode the labels and add them to the annotations
		data, err := json.Marshal(labels)
		if err != nil {
			return ocispec.Descriptor{}, fmt.Errorf("error encoding json labels: %w", err)
		}
		desc.Annotations[encoding.AnnotationLabels] = string(data)
	}

	if sourceIndex != nil {
		desc.Annotations[encoding.AnnotationSrcIndex] = string(sourceIndex)
	}
	return desc, nil
}
