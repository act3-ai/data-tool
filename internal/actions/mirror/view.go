package mirror

import (
	"context"
	// Register sha256/sha384/sha512 so go-digest can validate and verify
	// every algorithm the OCI image spec allows.
	_ "crypto/sha256"
	_ "crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"
)

const (
	kindIndex    = "index"
	kindImage    = "image"
	kindArtifact = "artifact"

	unnamedRef      = "<unnamed>"
	unknownPlatform = "unknown"

	mediaTypeDockerManifestList = "application/vnd.docker.distribution.manifest.list.v2+json"
	mediaTypeDockerManifest     = "application/vnd.docker.distribution.manifest.v2+json"
	mediaTypeDockerConfig       = "application/vnd.docker.container.image.v1+json"

	// Docker BuildKit marks attestation manifests inside an index with this
	// annotation (and platform unknown/unknown).
	annotationDockerReferenceType = "vnd.docker.reference.type"
)

// ErrIndexesDiffer is returned by --diff when the two indexes differ, so the
// command exits non-zero like diff(1).
var ErrIndexesDiffer = errors.New("OCI indexes differ")

// View lists or compares the contents of extracted OCI image layouts.
type View struct {
	*Action

	Detailed bool   // per-reference detail
	Refs     bool   // top-level digest, ref, kind
	Images   bool   // index/image/artifact rows, recursing into indexes
	Tree     bool   // tree rendering; combine with Refs (roots only) or Images (no blobs/referrers)
	Diff     string // compare top-level entries against this layout or index.json
}

func (action *View) validate() error {
	if action.Diff != "" && (action.Detailed || action.Refs || action.Images || action.Tree) {
		return errors.New("--diff cannot be combined with --detailed, --refs, --images, or --tree")
	}
	selected := 0
	for _, set := range []bool{action.Detailed, action.Refs, action.Images} {
		if set {
			selected++
		}
	}
	if selected > 1 {
		return errors.New("only one of --detailed, --refs, or --images may be used")
	}
	if action.Tree && action.Detailed {
		return errors.New("--tree cannot be combined with --detailed")
	}
	return nil
}

// Run renders the layout at layoutPath (a directory or its index.json).
// Machine-readable output goes to out; warnings about missing blobs go to errOut.
func (action *View) Run(ctx context.Context, out, errOut io.Writer, layoutPath string) error {
	if err := action.validate(); err != nil {
		return err
	}
	if action.Diff != "" {
		return diffIndexes(out, layoutPath, action.Diff)
	}

	storeDir, err := layoutDir(layoutPath)
	if err != nil {
		return err
	}
	index, err := readOCIIndex(storeDir)
	if err != nil {
		return err
	}

	// --refs reads index.json only, so it works on a bare gather index
	// that has no oci-layout file or blobs/ directory.
	if action.Refs {
		if action.Tree {
			return printRootLines(out, index)
		}
		return printRefs(out, index)
	}
	if len(index.Manifests) == 0 {
		if !action.Images {
			fmt.Fprintln(out, "No images or manifests found in index.json.")
		}
		return nil
	}

	// Open the layout read-only: nothing is extracted or written.
	store, err := oci.NewFromFS(ctx, os.DirFS(storeDir))
	if err != nil {
		return fmt.Errorf("opening OCI layout %s (this mode needs an extracted layout with oci-layout and blobs/; --refs and --diff accept a bare index.json): %w", storeDir, err)
	}
	v := &viewer{graph: store, errOut: errOut, color: useColor(out)}

	switch {
	case action.Tree:
		return v.printTrees(ctx, out, index, !action.Images)
	case action.Images:
		for _, desc := range index.Manifests {
			if err := v.printImages(ctx, out, desc); err != nil {
				return err
			}
		}
		return nil
	case action.Detailed:
		for _, desc := range index.Manifests {
			if err := v.printDetailed(ctx, out, desc); err != nil {
				return err
			}
		}
		return nil
	default:
		return v.printSummary(ctx, out, index)
	}
}

// ---------------------------------------------------------------------------
// index.json-level helpers (no blob access)
// ---------------------------------------------------------------------------

func layoutDir(p string) (string, error) {
	info, err := os.Stat(p)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return p, nil
	}
	if filepath.Base(p) == ocispec.ImageIndexFile {
		return filepath.Dir(p), nil
	}
	return "", fmt.Errorf("%s is not an extracted OCI layout directory or index.json; extract the archive before running mirror view", p)
}

func readOCIIndex(storeDir string) (ocispec.Index, error) {
	data, err := os.ReadFile(filepath.Join(storeDir, ocispec.ImageIndexFile))
	if err != nil {
		return ocispec.Index{}, fmt.Errorf("reading index.json: %w", err)
	}
	var index ocispec.Index
	if err := json.Unmarshal(data, &index); err != nil {
		return ocispec.Index{}, fmt.Errorf("decoding index.json: %w", err)
	}
	for _, desc := range index.Manifests {
		if err := desc.Digest.Validate(); err != nil {
			return ocispec.Index{}, fmt.Errorf("index.json: invalid digest %q: %w", desc.Digest, err)
		}
	}
	return index, nil
}

func referenceName(desc ocispec.Descriptor) string {
	if ref := desc.Annotations[ocispec.AnnotationRefName]; ref != "" {
		return ref
	}
	return unnamedRef
}

// descriptorKind classifies using only what the parent index says about a
// descriptor. --refs and --diff use it so their output does not depend on
// whether blobs are present.
func descriptorKind(desc ocispec.Descriptor) string {
	switch desc.MediaType {
	case ocispec.MediaTypeImageIndex, mediaTypeDockerManifestList:
		return kindIndex
	case ocispec.MediaTypeImageManifest, mediaTypeDockerManifest:
		if desc.ArtifactType != "" || desc.Annotations[annotationDockerReferenceType] != "" {
			return kindArtifact
		}
		return kindImage
	default:
		return kindArtifact
	}
}

func isManifest(mediaType string) bool {
	return mediaType == ocispec.MediaTypeImageManifest || mediaType == mediaTypeDockerManifest
}

// manifestKind refines descriptorKind using the manifest itself: OCI 1.1
// artifacts (SBOMs, signatures, Helm charts) share the image manifest media type.
func manifestKind(desc ocispec.Descriptor, m *ocispec.Manifest) string {
	kind := descriptorKind(desc)
	if kind != kindImage || m == nil {
		return kind
	}
	if m.ArtifactType != "" {
		return kindArtifact
	}
	switch m.Config.MediaType {
	case ocispec.MediaTypeImageConfig, mediaTypeDockerConfig:
		return kindImage
	default:
		return kindArtifact
	}
}

func artifactTypeOf(desc ocispec.Descriptor, m *ocispec.Manifest) string {
	switch {
	case m != nil && m.ArtifactType != "":
		return m.ArtifactType
	case desc.ArtifactType != "":
		return desc.ArtifactType
	case desc.Annotations[annotationDockerReferenceType] != "":
		return desc.Annotations[annotationDockerReferenceType]
	case m != nil && m.Config.MediaType != "":
		return m.Config.MediaType
	default:
		return desc.MediaType
	}
}

// formatPlatform renders os[(os.version)]/arch[/variant] as a single token.
func formatPlatform(p ocispec.Platform) string {
	if p.OS == "" || p.Architecture == "" {
		return ""
	}
	s := p.OS
	if p.OSVersion != "" {
		s += "(" + p.OSVersion + ")"
	}
	s += "/" + p.Architecture
	if p.Variant != "" {
		s += "/" + p.Variant
	}
	return s
}

func blobPath(d digest.Digest) string {
	return path.Join(ocispec.ImageBlobsDir, d.Algorithm().String(), d.Encoded())
}

func printRefs(w io.Writer, index ocispec.Index) error {
	for _, desc := range index.Manifests {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\n", desc.Digest, referenceName(desc), descriptorKind(desc)); err != nil {
			return err
		}
	}
	return nil
}

func printRootLines(w io.Writer, index ocispec.Index) error {
	color := useColor(w)
	for _, desc := range index.Manifests {
		fmt.Fprintf(w, "%s  %s\n", colorize(color, referenceName(desc)), treeDescriptor(desc))
	}
	return nil
}

// ---------------------------------------------------------------------------
// --diff
// ---------------------------------------------------------------------------

type indexEntry struct {
	digest digest.Digest
	ref    string
	kind   string
}

// diffIndexes compares the top-level descriptors of two index.json files.
// A row is identified by (digest, ref, kind), so a retag or an image→index
// change shows as one removal and one addition. Rows are sorted by ref with
// "-" before "+", keeping both sides of an update adjacent:
//
//	-<TAB>digest<TAB>ref<TAB>kind   only in againstPath
//	+<TAB>digest<TAB>ref<TAB>kind   only in inputPath
//
// Only index.json is read, so either side may be a bare gather index.
func diffIndexes(w io.Writer, inputPath, againstPath string) error {
	load := func(p string) (map[indexEntry]bool, error) {
		dir, err := layoutDir(p)
		if err != nil {
			return nil, err
		}
		index, err := readOCIIndex(dir)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		entries := make(map[indexEntry]bool, len(index.Manifests))
		for _, desc := range index.Manifests {
			entries[indexEntry{digest: desc.Digest, ref: referenceName(desc), kind: descriptorKind(desc)}] = true
		}
		return entries, nil
	}
	input, err := load(inputPath)
	if err != nil {
		return err
	}
	against, err := load(againstPath)
	if err != nil {
		return err
	}

	type row struct {
		sign string
		indexEntry
	}
	var rows []row
	for e := range against {
		if !input[e] {
			rows = append(rows, row{"-", e})
		}
	}
	for e := range input {
		if !against[e] {
			rows = append(rows, row{"+", e})
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if a.ref != b.ref {
			return a.ref < b.ref
		}
		if a.sign != b.sign {
			return a.sign == "-"
		}
		if a.digest != b.digest {
			return a.digest < b.digest
		}
		return a.kind < b.kind
	})
	for _, r := range rows {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.sign, r.digest, r.ref, r.kind)
	}
	if len(rows) > 0 {
		return ErrIndexesDiffer
	}
	return nil
}

// ---------------------------------------------------------------------------
// Blob-backed views
// ---------------------------------------------------------------------------

type viewer struct {
	graph  content.ReadOnlyGraphStorage
	errOut io.Writer
	color  bool
}

// node is a descriptor plus whatever could be read about it from the layout.
type node struct {
	desc         ocispec.Descriptor
	kind         string
	index        *ocispec.Index
	manifest     *ocispec.Manifest
	config       *ocispec.Image
	platform     string
	artifactType string
	missing      bool
}

func (n node) platformOrUnknown() string {
	if n.platform == "" {
		return unknownPlatform
	}
	return n.platform
}

// fetchJSON reads a blob through the store, which resolves
// blobs/<algorithm>/<encoded> from the digest and verifies size and digest.
func (v *viewer) fetchJSON(ctx context.Context, desc ocispec.Descriptor, target any) error {
	if err := desc.Digest.Validate(); err != nil {
		return fmt.Errorf("invalid digest %q: %w", desc.Digest, err)
	}
	data, err := content.FetchAll(ctx, v.graph, desc)
	if err != nil {
		return fmt.Errorf("reading %s: %w", blobPath(desc.Digest), err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decoding %s: %w", blobPath(desc.Digest), err)
	}
	return nil
}

// tolerateMissing turns a not-found error into a warning so partial or
// platform-filtered layouts can still be viewed. Other errors are returned.
func (v *viewer) tolerateMissing(err error) (missing bool, _ error) {
	if err == nil {
		return false, nil
	}
	if errors.Is(err, errdef.ErrNotFound) {
		fmt.Fprintf(v.errOut, "warning: %v\n", err)
		return true, nil
	}
	return false, err
}

func (v *viewer) inspect(ctx context.Context, desc ocispec.Descriptor) (node, error) {
	n := node{desc: desc, kind: descriptorKind(desc)}
	if desc.Platform != nil {
		n.platform = formatPlatform(*desc.Platform)
	}

	switch {
	case n.kind == kindIndex:
		var idx ocispec.Index
		missing, err := v.tolerateMissing(v.fetchJSON(ctx, desc, &idx))
		if err != nil {
			return n, err
		}
		n.missing = missing
		if !missing {
			n.index = &idx
		}
		return n, nil

	case isManifest(desc.MediaType):
		var m ocispec.Manifest
		missing, err := v.tolerateMissing(v.fetchJSON(ctx, desc, &m))
		if err != nil {
			return n, err
		}
		n.missing = missing
		if !missing {
			n.manifest = &m
			n.kind = manifestKind(desc, &m)
		}
		if n.kind == kindArtifact {
			n.artifactType = artifactTypeOf(desc, n.manifest)
			return n, nil
		}
		if n.manifest == nil || n.manifest.Config.Digest == "" {
			return n, nil
		}
		var cfg ocispec.Image
		missing, err = v.tolerateMissing(v.fetchJSON(ctx, n.manifest.Config, &cfg))
		if err != nil {
			return n, err
		}
		if !missing {
			n.config = &cfg
			if n.platform == "" {
				n.platform = formatPlatform(ocispec.Platform{
					OS: cfg.OS, OSVersion: cfg.OSVersion, Architecture: cfg.Architecture, Variant: cfg.Variant,
				})
			}
		}
		return n, nil

	default:
		n.artifactType = artifactTypeOf(desc, nil)
		return n, nil
	}
}

// --images rows; the second field selects the meaning of the rest:
//
//	digest<TAB>index<TAB>ref
//	digest<TAB>image<TAB>platform<TAB>ref
//	digest<TAB>artifact<TAB>artifactType<TAB>ref
//
// Children of an index follow their parent. They are untagged, so their ref
// is <unnamed> unless the index annotates them; they never inherit the
// parent's tag because that tag does not resolve to their digest.
func (v *viewer) printImages(ctx context.Context, w io.Writer, desc ocispec.Descriptor) error {
	n, err := v.inspect(ctx, desc)
	if err != nil {
		return err
	}
	ref := referenceName(desc)
	switch n.kind {
	case kindIndex:
		fmt.Fprintf(w, "%s\t%s\t%s\n", desc.Digest, kindIndex, ref)
		if n.index != nil {
			for _, child := range n.index.Manifests {
				if err := v.printImages(ctx, w, child); err != nil {
					return err
				}
			}
		}
	case kindImage:
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", desc.Digest, kindImage, n.platformOrUnknown(), ref)
	default:
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", desc.Digest, kindArtifact, n.artifactType, ref)
	}
	return nil
}

// printSummary is the default view: one row per top-level reference.
func (v *viewer) printSummary(ctx context.Context, out io.Writer, index ocispec.Index) error {
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	for _, desc := range index.Manifests {
		n, err := v.inspect(ctx, desc)
		if err != nil {
			return err
		}
		detail, err := v.summaryDetail(ctx, n)
		if err != nil {
			return err
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", desc.Digest, referenceName(desc), n.kind, detail)
	}
	return tw.Flush()
}

func (v *viewer) summaryDetail(ctx context.Context, n node) (string, error) {
	if n.missing {
		return "(blob missing: " + blobPath(n.desc.Digest) + ")", nil
	}
	switch n.kind {
	case kindIndex:
		var platforms []string
		artifacts := 0
		for _, child := range n.index.Manifests {
			c, err := v.inspect(ctx, child)
			if err != nil {
				return "", err
			}
			if c.kind == kindImage {
				platforms = append(platforms, c.platformOrUnknown())
			} else {
				artifacts++
			}
		}
		s := fmt.Sprintf("%d images", len(platforms))
		if len(platforms) > 0 {
			s += " (" + strings.Join(platforms, ", ") + ")"
		}
		if artifacts > 0 {
			s += fmt.Sprintf(", %d artifacts", artifacts)
		}
		return s, nil
	case kindImage:
		return fmt.Sprintf("%s  %d layers", n.platformOrUnknown(), layerCount(n)), nil
	default:
		return fmt.Sprintf("%s  %d layers", n.artifactType, layerCount(n)), nil
	}
}

func layerCount(n node) int {
	if n.manifest == nil {
		return 0
	}
	return len(n.manifest.Layers)
}

func detailField(w io.Writer, name, value string) {
	fmt.Fprintf(w, "%-15s%s\n", name+":", value)
}

// printDetailed writes every record, including error states, to the same
// writer in index order.
func (v *viewer) printDetailed(ctx context.Context, w io.Writer, desc ocispec.Descriptor) error {
	n, err := v.inspect(ctx, desc)
	if err != nil {
		return err
	}
	detailField(w, "Digest", desc.Digest.String())
	detailField(w, "Reference", referenceName(desc))
	detailField(w, "Kind", n.kind)
	detailField(w, "Media Type", desc.MediaType)
	detailField(w, "Size", treeSize(desc.Size))

	switch {
	case n.missing:
		detailField(w, "Status", "blob missing: "+blobPath(desc.Digest))
	case n.kind == kindIndex:
		fmt.Fprintf(w, "%-15s%d\n", "Manifests:", len(n.index.Manifests))
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		for _, child := range n.index.Manifests {
			c, err := v.inspect(ctx, child)
			if err != nil {
				return err
			}
			detail := c.platformOrUnknown()
			if c.kind == kindArtifact {
				detail = c.artifactType
			}
			if c.missing {
				detail += " (blob missing)"
			}
			fmt.Fprintf(tw, "  - %s\t%s\t%s\n", child.Digest, c.kind, detail)
		}
		if err := tw.Flush(); err != nil {
			return err
		}
	default:
		if n.kind == kindArtifact {
			detailField(w, "Artifact Type", n.artifactType)
		} else {
			detailField(w, "Platform", n.platformOrUnknown())
		}
		if n.config != nil && n.config.Created != nil && !n.config.Created.IsZero() {
			detailField(w, "Created", n.config.Created.UTC().Format("2006-01-02 15:04:05 MST"))
		}
		if n.manifest != nil {
			detailField(w, "Config Digest", n.manifest.Config.Digest.String())
			var total int64
			for _, l := range n.manifest.Layers {
				total += l.Size
			}
			detailField(w, "Layers", fmt.Sprintf("%d (%s)", len(n.manifest.Layers), treeSize(total)))
		}
	}

	keys := make([]string, 0, len(desc.Annotations))
	for k := range desc.Annotations {
		if k != ocispec.AnnotationRefName {
			keys = append(keys, k)
		}
	}
	if len(keys) > 0 {
		sort.Strings(keys)
		fmt.Fprintln(w, "Annotations:")
		for _, k := range keys {
			fmt.Fprintf(w, "  - %s: %s\n", k, desc.Annotations[k])
		}
	}
	_, err = fmt.Fprintln(w, strings.Repeat("-", 80))
	return err
}

// ---------------------------------------------------------------------------
// --tree
// ---------------------------------------------------------------------------

func useColor(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func colorize(enabled bool, s string) string {
	if !enabled {
		return s
	}
	return "\x1b[34m" + s + "\x1b[0m"
}

// treeSize uses decimal units, matching oci tree.
func treeSize(n int64) string {
	if n < 1000 {
		return strconv.FormatInt(n, 10) + " B"
	}
	value := float64(n)
	unit := ""
	for _, unit = range []string{"kB", "MB", "GB", "TB"} {
		value /= 1000
		if value < 1000 {
			break
		}
	}
	return fmt.Sprintf("%.1f %s", value, unit)
}

func treeDescriptor(desc ocispec.Descriptor) string {
	short := desc.Digest.String()
	if desc.Digest.Validate() == nil {
		short = desc.Digest.Encoded()
		if len(short) > 12 {
			short = short[:12]
		}
	}
	return fmt.Sprintf("[%-8s] %s %s", treeSize(desc.Size), desc.MediaType, short)
}

// treeSizes sums descriptor sizes. Sizes come from descriptors; layer
// contents are never read.
type treeSizes struct {
	total  int64
	unique int64
	seen   map[digest.Digest]bool
}

func newTreeSizes() *treeSizes { return &treeSizes{seen: make(map[digest.Digest]bool)} }

func (s *treeSizes) add(desc ocispec.Descriptor) {
	s.total += desc.Size
	if !s.seen[desc.Digest] {
		s.unique += desc.Size
		s.seen[desc.Digest] = true
	}
}

// treeLine is built first and rendered second, so connectors and
// continuation prefixes are computed in one place.
type treeLine struct {
	text     string
	children []*treeLine
}

func writeTree(w io.Writer, t *treeLine, prefix string) {
	for i, c := range t.children {
		branch, next := "├── ", "│   "
		if i == len(t.children)-1 {
			branch, next = "└── ", "    "
		}
		fmt.Fprintf(w, "%s%s%s\n", prefix, branch, c.text)
		writeTree(w, c, prefix+next)
	}
}

func (v *viewer) printTrees(ctx context.Context, out io.Writer, index ocispec.Index, full bool) error {
	global := newTreeSizes()
	for i, desc := range index.Manifests {
		if i > 0 {
			fmt.Fprintln(out)
		}
		local := newTreeSizes()
		root, err := v.buildTree(ctx, desc, full, roleRoot, local, global)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s  %s\n", colorize(v.color, referenceName(desc)), root.text)
		writeTree(out, root, "")
		fmt.Fprintf(out, "%s %s (%s deduplicated)\n", totalLabel(full), treeSize(local.total), treeSize(local.unique))
	}
	if len(index.Manifests) > 1 {
		fmt.Fprintf(out, "\nAll references %s %s (%s deduplicated)\n", strings.ToLower(totalLabel(full)), treeSize(global.total), treeSize(global.unique))
	}
	return nil
}

func totalLabel(full bool) string {
	if full {
		return "Total:"
	}
	return "Total (manifests only):"
}

// buildTree follows outgoing descriptors from desc. With full set it also
// shows config/layers and referrers (manifests whose subject is desc).
// Referrers are discovered through the store's predecessor graph, which ORAS
// builds from everything reachable from index.json, so a referrer that is not
// itself reachable from index.json will not be shown.
type treeRole int

const (
	roleRoot     treeRole = iota // ref is printed by the caller
	roleChild                    // entry of a parent index
	roleReferrer                 // manifest whose subject is the parent
)

func (v *viewer) buildTree(ctx context.Context, desc ocispec.Descriptor, full bool, role treeRole, local, global *treeSizes) (*treeLine, error) {
	n, err := v.inspect(ctx, desc)
	if err != nil {
		return nil, err
	}
	local.add(desc)
	global.add(desc)
	t := &treeLine{text: v.nodeLabel(n, role)}

	if n.index != nil {
		for _, child := range n.index.Manifests {
			c, err := v.buildTree(ctx, child, full, roleChild, local, global)
			if err != nil {
				return nil, err
			}
			t.children = append(t.children, c)
		}
	}

	if !full || n.missing {
		return t, nil
	}

	if m := n.manifest; m != nil {
		if m.Config.Digest != "" {
			local.add(m.Config)
			global.add(m.Config)
			t.children = append(t.children, &treeLine{
				text: "🗃️  " + colorize(v.color, "config") + "  " + treeDescriptor(m.Config),
			})
		}
		if len(m.Layers) > 0 {
			group := &treeLine{}
			var layerTotal int64
			for _, layer := range m.Layers {
				local.add(layer)
				global.add(layer)
				layerTotal += layer.Size
				group.children = append(group.children, &treeLine{text: treeDescriptor(layer)})
			}
			group.text = fmt.Sprintf("📦 %s  %d, %s total", colorize(v.color, "layers"), len(m.Layers), treeSize(layerTotal))
			t.children = append(t.children, group)
		}
	}

	referrers, err := registry.Referrers(ctx, v.graph, desc, "")
	if err != nil {
		return nil, fmt.Errorf("listing referrers of %s: %w", desc.Digest, err)
	}
	sort.Slice(referrers, func(i, j int) bool { return referrers[i].Digest < referrers[j].Digest })
	for _, r := range referrers {
		c, err := v.buildTree(ctx, r, full, roleReferrer, local, global)
		if err != nil {
			return nil, err
		}
		t.children = append(t.children, c)
	}
	return t, nil
}

func (v *viewer) nodeLabel(n node, role treeRole) string {
	var b strings.Builder
	switch {
	case role == roleReferrer:
		b.WriteString("🔗 " + colorize(v.color, "referrer") + "  " + n.artifactTypeOrMediaType())
	case n.kind == kindIndex:
		b.WriteString("📚 " + colorize(v.color, "index"))
	case n.kind == kindImage:
		b.WriteString("📜 " + colorize(v.color, "manifest") + "  " + n.platformOrUnknown())
	default:
		b.WriteString("📎 " + colorize(v.color, "artifact") + "  " + n.artifactType)
	}
	if ref := n.desc.Annotations[ocispec.AnnotationRefName]; ref != "" && role != roleRoot {
		b.WriteString("  " + ref)
	}
	b.WriteString("  " + treeDescriptor(n.desc))
	if n.missing {
		b.WriteString("  (blob missing)")
	}
	return b.String()
}

func (n node) artifactTypeOrMediaType() string {
	if n.artifactType != "" {
		return n.artifactType
	}
	return artifactTypeOf(n.desc, n.manifest)
}