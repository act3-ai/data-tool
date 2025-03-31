// Package print implements the pretty printing of OCI artifacts
package print

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"slices"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry"

	"gitlab.com/act3-ai/asce/data/schema/pkg/mediatype"
	"gitlab.com/act3-ai/asce/data/tool/internal/cache"
	"gitlab.com/act3-ai/asce/data/tool/internal/descriptor"
	"gitlab.com/act3-ai/asce/data/tool/internal/mirror/encoding"
)

// colors to print color to the console.
var (
	blue   = color.New(color.FgBlue).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	bold   = color.New(color.Bold).SprintFunc()
)

var (
	nameIndex        = "🗃 " + blue("index")     // aka image index, image list, manifest list
	nameManifest     = "📜 " + blue("manifest")  // aka image manifest
	nameManifests    = "🗂 " + blue("manifests") // manifests field of the index
	nameSubject      = "🤵 " + blue("subject")
	nameConfig       = "🎛  " + blue("config") // extra space needed here for some reason
	nameLayers       = "🥞 " + blue("layers")
	nameReferrers    = "👈 " + blue("referrers")
	namePredecessors = "👈 " + blue("predecessors")
)

var icons = map[string]string{
	ocispec.MediaTypeImageConfig:    "🖼 ",
	ocispec.MediaTypeImageIndex:     "🗃 ",
	ocispec.MediaTypeImageManifest:  "📜",
	ocispec.MediaTypeImageLayer:     "💾",
	ocispec.MediaTypeImageLayerGzip: "💾",
	ocispec.MediaTypeImageLayerZstd: "💾",
	mediatype.MediaTypeBottle:       "🍾", // ACE data bottle
	mediatype.MediaTypeBottleConfig: "🍾", // ACE data bottle config
}

// prefixes to print the tree branches.
const (
	prefixEmpty        = "   "
	prefixContinue     = "┃  "
	prefixItem         = "┣━ "
	prefixLast         = "┗━ "
	prefixContinueInfo = "│  "
	prefixItemInfo     = "├╴ "
	prefixLastInfo     = "└╴ "
)

// For more unicode characters in this class
// https://www.compart.com/en/unicode/block/U+2500

// Options for the printer.
type Options struct {
	Prefix string
	Depth  int

	DisableBlobs    bool
	ShortDigests    bool
	ShowAnnotations bool

	OnlyReferrers bool   // only used by predecessors
	ArtifactType  string // used when OnlyReferrers is true
}

// NOTE We could consider leveraging https://github.com/xlab/treeprint

// prettyPrinter is a printer to nicely format OCI artifacts in a tree structure.
type prettyPrinter struct {
	out io.Writer

	*Options

	bt            *cache.BytesTracker
	seenManifests map[descriptor.Descriptor]struct{}

	// current depth of the printer
	depth int

	// prefix to use for the first item
	prefix string

	// prefix to use for the remaining items
	prefixRest string
}

// All will print the successors an predecessors of the given descriptor.
func All(ctx context.Context, out io.Writer, storage content.ReadOnlyGraphStorage, node ocispec.Descriptor, options Options) error {
	pp := newPrinter(out, options)
	if _, err := fmt.Fprintln(out, bold("Successors of")); err != nil {
		return err
	}
	if err := pp.printSuccessors(ctx, storage, node); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(out, bold("\nPredecessors of")); err != nil {
		return err
	}
	return pp.printPredecessors(ctx, storage, node)
}

// Successors the successors of the given descriptor.
func Successors(ctx context.Context, out io.Writer, fetcher content.Fetcher, desc ocispec.Descriptor, options Options) error {
	pp := newPrinter(out, options)
	return pp.printSuccessors(ctx, fetcher, desc)
}

// Predecessors the successors of the given descriptor.
func Predecessors(ctx context.Context, out io.Writer, storage content.ReadOnlyGraphStorage, desc ocispec.Descriptor, options Options) error {
	pp := newPrinter(out, options)
	return pp.printPredecessors(ctx, storage, desc)
}

// newPrinter creates a new Pretty Printer.
func newPrinter(out io.Writer, options Options) *prettyPrinter {
	if options.Depth == 0 {
		options.Depth = 100 // math.MaxInt
	}

	return &prettyPrinter{
		out:           out,
		Options:       &options,
		bt:            &cache.BytesTracker{},
		seenManifests: map[descriptor.Descriptor]struct{}{},
		prefix:        options.Prefix,
		prefixRest:    options.Prefix,
	}
}

func (pp prettyPrinter) println(str string) error {
	_, err := fmt.Fprint(pp.out, pp.prefix, str, "\n")
	return err
}

// printSuccessors prints the successors only.
func (pp prettyPrinter) printSuccessors(ctx context.Context, fetcher content.Fetcher, node ocispec.Descriptor) error {
	size, err := pp.processDescriptor(ctx, fetcher, node)
	if err != nil {
		return err
	}

	if size != pp.bt.Total {
		panic(fmt.Sprintf("Expected the size %d to match the total size from the byte tracker %d", size, pp.bt.Total))
	}

	_, err = fmt.Fprintf(pp.out, "%sTotal: %s (%s deduplicated)\n", pp.prefix, Bytes(pp.bt.Total), Bytes(pp.bt.Deduplicated))
	return err
}

// printPredecessors prints the predecessors only.
func (pp prettyPrinter) printPredecessors(ctx context.Context, storage content.ReadOnlyGraphStorage, node ocispec.Descriptor) error {
	if err := pp.printDescriptor("", node); err != nil {
		return err
	}

	// display referrers (of any artifact type) as predecessors
	var predecessors []ocispec.Descriptor
	var name string
	if pp.OnlyReferrers {
		referrers, err := registry.Referrers(ctx, storage, node, pp.ArtifactType)
		if err != nil {
			return fmt.Errorf("getting referrers %v: %w", node, err)
		}
		predecessors = referrers
		name = nameReferrers
	} else {
		p, err := storage.Predecessors(ctx, node)
		if err != nil {
			return fmt.Errorf("getting predecessors %v: %w", node, err)
		}
		predecessors = p
		name = namePredecessors
	}

	if len(predecessors) == 0 {
		return nil
	}

	// The order is not deterministic from Predecessors or Referrers so we sort
	slices.SortFunc(predecessors, func(a, b ocispec.Descriptor) int {
		return cmp.Compare(a.Digest, b.Digest)
	})

	if err := pp.nest("").println(name); err != nil {
		return err
	}

	nested := pp.nest(prefixEmpty)

	n := len(predecessors)
	for i, d := range predecessors {
		last := i+1 == n
		child := nested.entry(last)

		if err := child.printPredecessors(ctx, storage, d); err != nil {
			return err
		}
	}

	return nil
}

// entry returns a new printer that prints primary items.  Last must be true for the last item in the list.
func (pp prettyPrinter) entry(last bool) prettyPrinter {
	prefix := pp.prefix
	if last {
		pp.prefix = prefix + prefixLast
		pp.prefixRest = prefix + prefixEmpty
	} else {
		pp.prefix = prefix + prefixItem
		pp.prefixRest = prefix + prefixContinue
	}
	return pp
}

// info returns a new printer that prints info items.  Last must be true for the last item in the list.
//
//nolint:unused
func (pp prettyPrinter) info(last bool) prettyPrinter {
	prefix := pp.prefix
	if last {
		pp.prefix = prefix + prefixLastInfo
		pp.prefixRest = prefix + prefixEmpty
	} else {
		pp.prefix = prefix + prefixItemInfo
		pp.prefixRest = prefix + prefixContinueInfo
	}
	return pp
}

// nest returns a new printer that prints nested one level deeper with the addition prefix appended.
func (pp prettyPrinter) nest(prefix string) prettyPrinter {
	pp.prefix = pp.prefixRest + prefix
	pp.prefixRest += prefix
	return pp
}

// processDescriptor determines the appropriate printing func by descriptor media type.
// If storage implements content.PredecessorFinder then we also display predecessors.
//
//nolint:revive
func (pp prettyPrinter) processDescriptor(ctx context.Context, fetcher content.Fetcher, desc ocispec.Descriptor) (int64, error) {
	// check for max depth
	if pp.depth >= pp.Depth {
		return 0, pp.println("pruned (max depth reached)")
	}
	pp.depth++

	if err := pp.printDescriptor("", desc); err != nil {
		return 0, err
	}

	var size int64
	switch {
	case encoding.IsIndex(desc.MediaType):
		n, err := pp.printImageIndex(ctx, fetcher, desc)
		if err != nil {
			return 0, err
		}
		size += n
	case encoding.IsImage(desc.MediaType):
		n, err := pp.printImageManifest(ctx, fetcher, desc)
		if err != nil {
			return 0, err
		}
		size += n
	default:
		return desc.Size, nil
	}

	return size, nil
}

// prettyPrintImageIndex calls prettyPrintDescriptor for all indicies.
func (pp prettyPrinter) printImageIndex(ctx context.Context, fetcher content.Fetcher, desc ocispec.Descriptor) (int64, error) {
	// check for duplicates
	d := descriptor.FromOCI(desc)
	if _, ok := pp.seenManifests[d]; ok {
		// we have already processed this descriptor
		return desc.Size, pp.nest("").println("(duplicate suppressed)")
	}
	pp.seenManifests[d] = struct{}{}

	sum := desc.Size

	// cannot use encoding.Successors() here because we need to distinguish between the subject, manifest, and extra manifests
	manifest := &ocispec.Index{}
	manifestBytes, err := content.FetchAll(ctx, fetcher, desc)
	if err != nil {
		return 0, fmt.Errorf("fetching manifest: %w", err)
	}
	err = json.Unmarshal(manifestBytes, manifest)
	if err != nil {
		return 0, fmt.Errorf("parsing manifest: %w", err)
	}

	if err := pp.nest("").printHeader(nameIndex, manifest.Annotations); err != nil {
		return 0, err
	}

	extraManifests, err := encoding.ExtraManifests(manifest)
	if err != nil {
		return 0, err
	}
	// TODO differentiate between manifests and extra manifests
	manifest.Manifests = append(manifest.Manifests, extraManifests...)

	info := pp.nest(prefixContinueInfo)

	manifests := info.entry(manifest.Subject == nil)
	if err := manifests.println(nameManifests); err != nil {
		return 0, err
	}

	manifestsNested := manifests.nest(prefixEmpty)

	n := len(manifest.Manifests)
	for i, desc := range manifest.Manifests {
		last := i+1 == n
		child := manifestsNested.entry(last)

		n, err := child.processDescriptor(ctx, fetcher, desc)
		if err != nil {
			return 0, err
		}
		sum += n
	}

	if manifest.Subject != nil {
		s := info.entry(true)
		if err := s.println(nameSubject); err != nil {
			return 0, err
		}

		entry := s.nest(prefixEmpty)

		n, err := entry.processDescriptor(ctx, fetcher, *manifest.Subject)
		if err != nil {
			return 0, err
		}
		sum += n
	}

	return sum, pp.nest(prefixLastInfo).printSummary(sum)
}

// printImageManifest prints the image manifest and its parts.
func (pp prettyPrinter) printImageManifest(ctx context.Context, fetcher content.Fetcher, desc ocispec.Descriptor) (int64, error) {
	// check for duplicates
	d := descriptor.FromOCI(desc)
	if _, ok := pp.seenManifests[d]; ok {
		// we have already processed this descriptor
		return desc.Size, pp.nest("").println("(duplicate suppressed)")
	}
	pp.seenManifests[d] = struct{}{}

	sum := desc.Size

	// cannot use encoding.Successors() here because we need to distinguish between the subject, config, and layers
	manifest := &ocispec.Manifest{}
	manifestBytes, err := content.FetchAll(ctx, fetcher, desc)
	if err != nil {
		return 0, fmt.Errorf("fetching manifest: %w", err)
	}

	if err := json.Unmarshal(manifestBytes, manifest); err != nil {
		return 0, fmt.Errorf("parsing manifest: %w", err)
	}

	if err := pp.nest("").printHeader(nameManifest, manifest.Annotations); err != nil {
		return 0, err
	}

	info := pp.nest(prefixContinueInfo)

	sum += manifest.Config.Size
	if err := info.entry(false).printDescriptor(nameConfig, manifest.Config); err != nil {
		return 0, err
	}

	layers := info.entry(manifest.Subject == nil)
	if err := layers.println(nameLayers); err != nil {
		return 0, err
	}

	layersNested := layers.nest(prefixEmpty)
	n := len(manifest.Layers)
	for i, desc := range manifest.Layers {
		sum += desc.Size
		last := i+1 == n
		entry := layersNested.entry(last)
		if err := entry.printDescriptor("", desc); err != nil {
			return 0, err
		}
	}

	if manifest.Subject != nil {
		s := info.entry(true)
		if err := s.println(nameSubject); err != nil {
			return 0, err
		}

		entry := s.nest(prefixEmpty)

		n, err := entry.processDescriptor(ctx, fetcher, *manifest.Subject)
		if err != nil {
			return 0, err
		}
		sum += n
	}

	return sum, pp.nest(prefixLastInfo).printSummary(sum)
}

func (pp prettyPrinter) printDescriptor(itemName string, desc ocispec.Descriptor) error {
	pp.bt.Add(desc)

	if pp.DisableBlobs && !encoding.IsManifest(desc.MediaType) {
		return nil
	}

	size := Bytes(desc.Size)

	var artifactType string
	if desc.ArtifactType != "" {
		artifactType = fmt.Sprintf(" (%s)", formatMediaType(desc.ArtifactType))
	}

	var platform string
	if desc.Platform != nil {
		if desc.Platform.OS != "unknown" {
			platform = fmt.Sprintf(" (%s)", formatMediaType(desc.Platform.OS+"/"+desc.Platform.Architecture+"/"+desc.Platform.Variant))
		}
	}

	var dgst string
	if pp.ShortDigests {
		dgst = ShortDigest(desc.Digest)
	} else {
		dgst = desc.Digest.String()
	}

	var annos string
	if pp.ShowAnnotations {
		annos = formatAnnotations(desc.Annotations)
	}

	_, err := fmt.Fprintf(pp.out, "%s%s[%6s] %s%s%s %s %s\n", pp.prefix, itemName, size, formatMediaType(desc.MediaType), artifactType, platform, dgst, annos)
	return err
}

func (pp prettyPrinter) printHeader(name string, annotations map[string]string) error {
	var annos string
	if pp.ShowAnnotations {
		annos = formatAnnotations(annotations)
	}

	_, err := fmt.Fprintf(pp.out, "%s%s %s\n", pp.prefix, name, annos)
	return err
}

func (pp prettyPrinter) printSummary(sum int64) error {
	_, err := fmt.Fprintf(pp.out, "%s%s total\n", pp.prefix, Bytes(sum))
	return err
}

func formatMediaType(mt string) string {
	if icon := icons[mt]; icon != "" {
		return icon + " " + mt
	}
	return mt
}

func formatAnnotations(annotations map[string]string) string {
	var annos string
	for k, v := range annotations {
		// TODO escape newlines from values and possibly omit if too long.
		// maybe JSON encode the values to escape them?
		annos += fmt.Sprintf("%s=%.20s ", k, v)
	}

	if len(annos) != 0 {
		annos = yellow(annos)
	}
	return annos
}

// ShortDigest returns a compact representation of a digest (dropping information for brevity).
func ShortDigest(d digest.Digest) string {
	return d.Encoded()[:12]
}

// Bytes formats the bytes in a human readable way.  It must be positive.
func Bytes(bytes int64) string {
	if bytes < 0 {
		panic(fmt.Sprintf("Bytes must be positive (got %d)", bytes))
	}
	return humanize.Bytes(uint64(bytes))
}

// TODO add parsing of media types to display the item differently
// TODO verbosity levels (depth levels)
