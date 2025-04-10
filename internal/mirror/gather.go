package mirror

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/sync/errgroup"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"

	"github.com/act3-ai/data-tool/internal/mirror/encoding"
	"github.com/act3-ai/data-tool/internal/print"
	"github.com/act3-ai/data-tool/internal/ref"
	dtreg "github.com/act3-ai/data-tool/internal/registry"
	"github.com/act3-ai/data-tool/internal/ui"
	reg "github.com/act3-ai/data-tool/pkg/registry"
	"github.com/act3-ai/go-common/pkg/logger"
)

// GatherOptions specify the requirements to run a mirror gather operation.
type GatherOptions struct {
	Platforms      []string
	ConcurrentHTTP int
	DestStorage    content.GraphStorage
	Log            *slog.Logger
	RootUI         *ui.Task
	SourceFile     string
	Dest           string
	Annotations    map[string]string
	IndexFallback  bool
	DestReference  registry.Reference
	Recursive      bool
	Targeter       reg.GraphTargeter
}

// Gather will take the references defined in a SourceFile and consolidate them to a destination target.
func Gather(ctx context.Context, dataToolVersion string, opts GatherOptions) (ocispec.Descriptor, error) { //nolint:gocognit
	// throw the platforms in a map for easy querying
	var platforms []*ocispec.Platform
	if len(opts.Platforms) != 0 {
		plat, err := parsePlatforms(opts.Platforms)
		if err != nil {
			return ocispec.Descriptor{}, fmt.Errorf("error parsing the platforms: %w", err)
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
	sourceList, err := ProcessSourcesFile(ctx, opts.SourceFile, nil, opts.ConcurrentHTTP)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	var i int
	for _, src := range sourceList {
		i++
		task := opts.RootUI.SubTask(fmt.Sprintf("Source %d", i))
		g.Go(func() error {
			defer task.Complete()
			task.Infof("Copying %s", src.Name)
			log := logger.V(opts.Log, 1)
			log.InfoContext(gctx, "copying", "srcReference", src.Name, "dest reference", opts.Dest)
			// TODO: add progress back in... use GraphCopyOptions pre and post manifest
			numBytes := atomic.Int64{}

			srcTarget, err := opts.Targeter.GraphTarget(ctx, src.Name)
			if err != nil {
				return fmt.Errorf("initializing destination graph target: %w", err)
			}

			desc, err := srcTarget.Resolve(ctx, src.Name)
			if err != nil {
				return fmt.Errorf("resolving source descriptor '%s': %w", src.Name, err)
			}

			// resolve the endpoint if necessary
			srcRef, err := dtreg.ParseEndpointOrDefault(opts.Targeter, src.Name)
			if err != nil {
				return err
			}

			copyOpts := oras.CopyGraphOptions{
				MountFrom: mountFrom(srcRef, opts.DestReference),
				OnMounted: onMounted(opts.Log),
			}
			c, err := NewCopier(ctx, opts.Log, srcTarget, opts.DestStorage, desc, opts.Recursive, platforms, copyOpts)

			if err != nil {
				return err
			}

			// record the bytes and number of blobs that were actually copied.
			c.options.PostCopy = func(ctx context.Context, desc ocispec.Descriptor) error {
				// copied bytes and blobs
				numBytes.Add(desc.Size)
				wt.Add(desc)
				return nil
			}

			var descriptors []ocispec.Descriptor
			if platforms == nil {
				if err := Copy(gctx, c); err != nil {
					return fmt.Errorf("copying from %s: %w", src.Name, err)
				}
				desc, err := annotateManifest(src.Name, c.root, src.Labels, nil)
				if err != nil {
					return err
				}
				// count bytes
				if err := extractBlobs(ctx, bt.AddDescriptor, opts.DestStorage, desc); err != nil {
					return fmt.Errorf("counting bytes: %w", err)
				}
				descriptors = append(descriptors, desc)
			} else {
				platformDescriptors, err := CopyFilterOnPlatform(gctx, c)
				if err != nil {
					return fmt.Errorf("copying with specific platforms for source %s: %w", src.Name, err)
				}
				for _, d := range platformDescriptors {
					d, err := annotateManifest(src.Name, d, src.Labels, c.originatingIndex)
					if err != nil {
						return err
					}
					// count bytes
					if err := extractBlobs(ctx, bt.AddDescriptor, opts.DestStorage, d); err != nil {
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
		return ocispec.Descriptor{}, err
	}

	// set the ace-dt version, size, and deduplicated annotations
	opts.Annotations[encoding.AnnotationGatherVersion] = dataToolVersion
	opts.Annotations[encoding.AnnotationLayerSizeTotal] = fmt.Sprint(bt.Total)
	opts.Annotations[encoding.AnnotationLayerSizeDeduplicated] = fmt.Sprint(bt.Deduplicated)

	// sort based on the 'vnd.act3-ace.manifest.source' annotation, an effort to
	// improve readability by grouping registries together; e.g. all docker.io
	// refs will exist in the index consecutively.
	// if the annotation DNE, unlikely since we add this ourselves, put it at the end.
	slices.SortFunc(manifests, func(a, b ocispec.Descriptor) int {
		aAnnos, ok := a.Annotations[ref.AnnotationSrcRef]
		if !ok {
			aAnnos = "zz" // ensure "zot..."" < "zz"
		}
		bAnnos, ok := b.Annotations[ref.AnnotationSrcRef]
		if !ok {
			bAnnos = "zz"
		}
		return strings.Compare(aAnnos, bAnnos)
	})

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
		return ocispec.Descriptor{}, fmt.Errorf("marshalling the index: %w", err)
	}

	idxDesc, err := oras.PushBytes(ctx, opts.DestStorage, ocispec.MediaTypeImageIndex, b)
	switch {
	case err != nil && !errors.Is(err, errdef.ErrAlreadyExists):
		return ocispec.Descriptor{}, fmt.Errorf("pushing gather index manifest: %w", err)
	case errors.Is(err, errdef.ErrAlreadyExists):
		// in the case of directly using the local cache (e.g. in archive) we can recover
		idxDesc = ocispec.Descriptor{
			MediaType: ocispec.MediaTypeImageIndex,
			// ArtifactType: encoding.MediaTypeGather, // not necessary
			Digest: digest.FromBytes(b),
			Size:   int64(len(b)),
		}
		fallthrough
	default:
		// not deferring this log b/c it shouldn't display if gather fails
		opts.RootUI.Infof("Gathered %s (representing %s)", print.Bytes(bt.Deduplicated), print.Bytes(bt.Total))
		// small helper function to print out the minimum number of bytes that were copied over
		minBytesTransferred := func(a, b int64) int64 {
			if a < b {
				return a
			}
			return b
		}
		opts.RootUI.Infof("%s pushed for %d blobs", print.Bytes(minBytesTransferred(wt.transferred.Load(), bt.Deduplicated)), wt.blobs.Load())
	}

	return idxDesc, nil
}

func annotateManifest(srcRef string, desc ocispec.Descriptor, labels map[string]string, sourceIndex []byte) (ocispec.Descriptor, error) {
	desc.Annotations = map[string]string{
		ref.AnnotationSrcRef:      srcRef,
		ocispec.AnnotationRefName: srcRef,
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
