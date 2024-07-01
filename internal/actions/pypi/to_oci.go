package pypi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sourcegraph/conc/iter"
	"github.com/sourcegraph/conc/pool"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"

	"git.act3-ace.com/ace/data/schema/pkg/selectors"
	"git.act3-ace.com/ace/go-common/pkg/logger"
	"git.act3-ace.com/ace/go-common/pkg/redact"
	"gitlab.com/act3-ai/asce/data/tool/internal/python"
	"gitlab.com/act3-ai/asce/data/tool/internal/ui"
)

// Some motivation for this approach is given in https://stevelasker.blog/2021/08/26/artifact-services/

const (
	// PythonDistributionDigest is the annotation key used for denoting the digest (e.g., sha256:deedbef...) of the distribution file (often a whl).
	PythonDistributionDigest = "vnd.act3-ace.python.distribution.digest"

	// PythonDistributionFilename is the annotation key used for denoting the file name of the distribution file (often a mypkg-1.0.0-something.whl).
	PythonDistributionFilename = "vnd.act3-ace.python.distribution.filename"

	// PythonDistributionMetadataDigest is the annotation key used for denoting the digest of the METADATA of the whl or sdist.
	PythonDistributionMetadataDigest = "vnd.act3-ace.python.distribution.core-metadata.digest"

	// PythonDistributionVersion is the annotation key used for denoting the version of the distribution file (e.g., 1.2.3).
	PythonDistributionVersion = "vnd.act3-ace.python.distribution.version"

	// PythonDistributionRequiresPython is the annotation key used for denoting the requied python version to use the distribution file (e.g., >=3.7).
	PythonDistributionRequiresPython = "vnd.act3-ace.python.distribution.requires-python"

	// PythonDistributionProject is the annotation key used for denoting the project name of the distribution file (e.g., pyzmq).
	PythonDistributionProject = "vnd.act3-ace.python.distribution.project"

	// PythonDistributionYanked is the annotation key used for denoting that a distribution has been marked as yanked.  The value is the reason.  If this annotation is present with an empty value it means it is yanked but without a reason.
	PythonDistributionYanked = "vnd.act3-ace.python.distribution.yanked"

	// PythonSyncVersion is the annotation key used for denoting the sync software protocol/version used to create the manifest.
	PythonSyncVersion = "vnd.act3-ace.python.sync.version"
)

// MediaTypes for python artifacts.
const (
	MediaTypePythonDistributionWheel   = "applications/vnd.act3-ace.python.bdist.wheel+zip"
	MediaTypePythonDistributionSource  = "applications/vnd.act3-ace.python.sdist.tar+gzip"
	MediaTypePythonDistributionEgg     = "applications/vnd.act3-ace.python.bdist.egg+zip"
	MediaTypePythonDistributionUnknown = "applications/vnd.act3-ace.python.unknown"

	MediaTypePythonSourceDistribution   = "applications/vnd.act3-ace.python.sdist.tar+gzip"
	MediaTypePythonDistributionMetadata = "application/vnd.act3-ace.python.core-metadata.v1"
	MediaTypePythonDistributionIndex    = "application/vnd.act3-ace.python.index.v1+json"
	MediaTypePythonDistribution         = "application/vnd.act3-ace.python.distribution.v1+json"
)

// ToOCI represents the pypi to-oci action.
type ToOCI struct {
	*Action

	// Selectors holds the python/abi/platform selectors
	Selectors []string

	// DryRun only figures what work it would do.  It does not make any changes to OCI.
	DryRun bool

	// Reproducible removes timestamps.
	Reproducible bool

	// Clean will copy everything over assuming there is no existing OCI artifacts.  When true, the prior files that were downloaded are ignored when constructing artifacts.
	Clean bool

	// Show labels of the packages (to make is easier to find filters that work)
	// ShowLabels bool

	// IndexURL is the primary index URL
	IndexURL string

	// ExtraIndexURLs is a slice of extra index URLs to use to find packages.  These are processed after IndexURL.
	ExtraIndexURLs []string

	// RequirementFile is a slice of requirement files to sync
	RequirementFiles []string

	// Also include the extra packages from requirements
	IncludeExtras bool

	// Continue even if an error occurs in processing a Python project or distribution.
	// This is useful when combined with FailedRequirementsFile to collect the failures in a way that can be used to process the erroneous entries.
	ContinueOnError bool

	// FailedRequirementsPath is a file path to write the requirements that failed to be transferred
	FailedRequirementsFile string
}

// Run performs the pypi to-oci operation.
func (action *ToOCI) Run(ctx context.Context, repository string, additionalRequirements ...string) error {
	log := logger.FromContext(ctx)

	repo, err := action.Config.Repository(ctx, repository)
	if err != nil {
		return err
	}

	cfg := action.Config.Get(ctx)

	client, err := action.PyPIClient()
	if err != nil {
		return fmt.Errorf("getting PyPI HTTP auth client: %w", err)
	}

	opts := optionsToOCI{
		dryRun:       action.DryRun,
		allowYanked:  action.AllowYanked,
		reproducible: action.Reproducible,
		clean:        action.Clean,

		client:     client,
		repository: repo,
	}

	// get the selectors
	if len(action.Selectors) != 0 {
		s, err := selectors.Parse(action.Selectors)
		if err != nil {
			return err
		}
		opts.sels = s
	}

	requirements := python.Requirements{
		IndexURL:       action.IndexURL,
		ExtraIndexURLs: action.ExtraIndexURLs,
	}

	// collect requirements from the command line arguments
	for _, reqStr := range additionalRequirements {
		err := requirements.AddRequirementString(reqStr)
		if err != nil {
			return err
		}
	}

	// collect requirements from files
	for _, requirementFile := range action.RequirementFiles {
		err := requirements.ParseRequirementsFile(requirementFile)
		if err != nil {
			return err
		}
	}

	if action.IncludeExtras {
		if err := requirements.IncludeExtras(); err != nil {
			return fmt.Errorf("including extra packages: %w", err)
		}
	}

	opts.pypis = requirements.Indexes()

	// if log.Enabled() {
	for _, pypi := range opts.pypis {
		log.InfoContext(ctx, "Using Python package index", "pypi", redact.URLString(pypi))
	}
	//}

	// parallelize the embarrassingly parallel loop
	concurrency := max(cfg.ConcurrentHTTP/2, 1)
	opts.concurrency = max(cfg.ConcurrentHTTP - concurrency)
	logger.V(log, 1).InfoContext(ctx, "Looping over projects", "concurrency", concurrency)

	p := pool.New().
		WithMaxGoroutines(concurrency).
		WithContext(ctx)
	if !action.ContinueOnError {
		p = p.WithCancelOnError().WithFirstError()
	}

	projects := requirements.Projects()
	errs := make([]error, len(projects))
	for i, project := range projects {
		p.Go(func(ctx context.Context) error {
			var err error
			select {
			case <-ctx.Done():
				// bail out early if the context is already done
				err = ctx.Err()
			default:
				_, err = processRequirement(ctx, requirements.RequirementsForProject(project), opts)
			}
			// record the error
			errs[i] = err
			return err
		})
	}

	// TODO this does not handle the case where an error occurs.
	// It still calls processRequirement() on all subsequent false even if action.ContinueOnError is false.

	if err := p.Wait(); err != nil {
		if action.FailedRequirementsFile != "" {
			// TODO It would better if we could write this out as we processes the files.
			// This would ensure that the output is written even if the above code panics.

			// We would need to preserve the original order even though there is concurrency in the loop.
			// github.com/sourcegraph/conc/stream might be useful for this (concurrency with sequential callbacks)
			// but it does not support ctx arguments or error returns

			if err := writeFailedRequirements(action.FailedRequirementsFile, &requirements, projects, errs); err != nil {
				return err
			}

			return err
		}
		return err
	}

	return nil
}

func writeFailedRequirements(filename string, requirements *python.Requirements,
	projects []string, errs []error,
) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("creating failed requirements file: %w", err)
	}
	defer f.Close()

	if _, err := fmt.Fprintln(f, "--index-url", requirements.IndexURL); err != nil {
		return err
	}
	for _, s := range requirements.ExtraIndexURLs {
		if _, err := fmt.Fprintln(f, "--extra-index-url", s); err != nil {
			return err
		}
	}

	for i, project := range projects {
		e := errs[i]
		if e == nil {
			if _, err := fmt.Fprintf(f, "# %q processed successfully\n", project); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(f, "# %q failed to process with error %s\n", project, e); err != nil {
				return err
			}
			for _, req := range requirements.RequirementsForProject(project) {
				if _, err := fmt.Fprintln(f, req.String()); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

type optionsToOCI struct {
	dryRun, allowYanked, reproducible, clean bool

	client     remote.Client
	repository registry.Repository

	concurrency int

	sels  selectors.LabelSelectorSet
	pypis []string
}

func processRequirement(ctx context.Context, reqs []python.Requirement, opts optionsToOCI) (ocispec.Descriptor, error) {
	project := reqs[0].Name
	log := logger.FromContext(ctx).With("project", project)

	projectTask := ui.FromContextOrNoop(ctx).SubTask(project)
	defer projectTask.Complete()

	projectTask.Infof("Processing")

	var filtered []python.DistributionEntry
	var pyDistIdx *pythonDistributionIndex

	p := pool.New().WithContext(ctx)
	p.Go(func(ctx context.Context) error {
		// find the package on the package index
		dists, err := python.RetrieveAllDistributions(ctx, opts.client, opts.pypis, project)
		if err != nil {
			return err
		}

		f, err := filterDistributionEntries(log, dists, opts.sels, opts.allowYanked, reqs)
		if err != nil {
			return err
		}

		filtered = f
		return nil
	})

	p.Go(func(ctx context.Context) error {
		// Get the old Image Index if one exists
		idx, err := newPythonDistributionIndex(ctx, opts.repository, opts.clean, project)
		if err != nil {
			return err
		}

		pyDistIdx = idx
		return nil
	})

	if err := p.Wait(); err != nil {
		return ocispec.Descriptor{}, err
	}

	// NOTE from a data model perspective we can store a ImageIndex (for a project) that points to an ImageIndex (for each version) that points to an Image with a single layer for each distribution.  We need to determine if there is value in doing so.

	// parallelizing this loop
	logger.V(log, 1).InfoContext(ctx, "Looping over entries", "concurrency", opts.concurrency, "count", len(filtered))
	mapper := iter.Mapper[python.DistributionEntry, ocispec.Descriptor]{
		MaxGoroutines: opts.concurrency,
	}
	items, err := mapper.MapErr(filtered, func(entry *python.DistributionEntry) (ocispec.Descriptor, error) {
		return processEntry(ctx, projectTask, pyDistIdx, *entry, opts)
	})
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("processing entries failed: %w", err)
	}

	/*
		FIXME The below operation is not atomic.  If multiple processes are running the sync process then we can loose data (it will be recovered if re-run manually).
		Ideally we would use the referrers API in OCI spec if the registry supports it.
		If not then we can use the fallback referrers API approach with tags that are the digest.
		See https://github.com/opencontainers/distribution-spec/blob/main/spec.md#unavailable-referrers-api
		In our case I think we might only need to deal with the ETags with If-Match header to prevent mid-air collisions.
		See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/ETag
		For docker distribution the ETag is just the sha256 hash of the contents but we should probably not assume anything about the ETag value since that can change for a registry.
		1) To prevent mid-air collisions we can grab the ETag from the response to the GET manifest.
		2) When making the PUT/POST request to update the manifest we add an If-Match header with the same value.
		3) If the push fails with a 412 Precondition Failed then we start over by pulling the manifest down again and pushing it up again.  Maybe with a random delay.

		Crane has an issue for this https://github.com/google/go-containerregistry/issues/1333

		The Referrers API can be used to solve this problem as well.  We have to restructure how we store things in OCI.  We need to add a project level manifest (can be empty) that is tagged with the project name.  We already have one manifest per distribution file/entry but we set the subject field to point to the project manifest.  Then we can use the referrers API to lookup referrers of our project tag/manifest to get all the index of manifests (distribution files) for our project.
	*/

	for _, item := range items {
		pyDistIdx.Update(item)
	}

	// TODO we need to update the index so long as any changes are made (if one processEntry succeeded then we need to process it)
	indexData, err := pyDistIdx.IndexData()
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	var target oras.Target = opts.repository
	if opts.dryRun {
		target = memory.New()
	}

	log.InfoContext(ctx, "Writing image index")
	// TODO oras v2.6.0 is planning on having a oras.PackIndex()
	// see https://github.com/oras-project/oras-go/issues/576
	d, err := oras.TagBytes(ctx, target, ocispec.MediaTypeImageIndex, indexData, project)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("error writing index: %w", err)
	}
	indexDescriptor := d

	// TODO maybe use a callback to record this information.
	// repoStr := opts.repository.Reference.String()
	// repoTag := fmt.Sprint(repoStr, ":", project)
	// repoDigest := fmt.Sprint(repoStr, "@", indexDescriptor.Digest)
	// projectTask.Infof("Image index available at:\n\t%s\n\t%s", repoTag, repoDigest)

	log.InfoContext(ctx, "Done processed project", "indexDigest", indexDescriptor.Digest)
	return indexDescriptor, nil
}

// processEntry processes a single distribution entry by upload (if necessary) to OCI.
// The descriptor of the Manifest is used.
func processEntry(ctx context.Context, task *ui.Task,
	pyDistIdx *pythonDistributionIndex, entry python.DistributionEntry, opts optionsToOCI,
) (ocispec.Descriptor, error) {
	subUI := task.SubTask(entry.Filename)
	defer subUI.Complete()

	// we store the python distribution file digest and name the image index (or manifests or configs) as annotations so we can lookup them up by version (tag).

	standing := ""
	if entry.Yanked != nil {
		standing = fmt.Sprintf(" (warning yanked: %s)", *entry.Yanked)
	}

	dryRun := ""
	if opts.dryRun {
		dryRun = "(dry run)"
	}

	if man, ok := pyDistIdx.LookupDistribution(entry.Filename); ok {
		// cache hit, return the descriptor of the manifest
		subUI.Infof("Distribution exists, skipping %s", standing)
		return man, nil
	}
	subUI.Infof("Distribution missing %s, transferring %s", standing, dryRun)

	subUI.Info("Transferring distribution file")
	var blobTarget interface {
		content.Resolver
		content.Pusher
	} = opts.repository.Blobs()
	if opts.dryRun {
		blobTarget = memory.New()
	}
	layer, metadata, err := transferEntry(ctx, entry, opts.client, blobTarget)
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	// update the manifest (the layers are already there)
	layers := []ocispec.Descriptor{layer}
	if metadata != nil {
		// We store the metadata as a layer instead of the config for better compatibility with non-compliant registries.
		layers = append(layers, *metadata)
	}
	options := oras.PackManifestOptions{
		Layers:              layers,
		ManifestAnnotations: annotationsFromDistribution(entry, opts.reproducible),
	}

	subUI.Info("Finalizing distribution")
	var pusher content.Pusher = opts.repository
	if opts.dryRun {
		// create a place for PackManifest to dump the data.
		pusher = memory.New()
	}

	return oras.PackManifest(ctx, pusher, oras.PackManifestVersion1_1, MediaTypePythonDistribution, options) //nolint:wrapcheck
}

// transferEntry copies the distribution file and it's metadata returning those two descriptors, respectively.
func transferEntry(ctx context.Context, entry python.DistributionEntry, client remote.Client,
	target interface {
		content.Resolver
		content.Pusher
	},
) (ocispec.Descriptor, *ocispec.Descriptor, error) {
	var metadata *ocispec.Descriptor
	if metadataURL := entry.MetadataURL(); metadataURL != "" {
		meta, err := transferBlob(ctx, metadataURL, entry.MetadataDigest, MediaTypePythonDistributionMetadata, client, target)
		if err != nil {
			return ocispec.Descriptor{}, nil, err
		}
		metadata = &meta
	}

	// find the media type for the actual data
	distMediaType := MediaTypePythonDistributionUnknown
	switch {
	case strings.HasSuffix(entry.Filename, ".whl"):
		distMediaType = MediaTypePythonDistributionWheel
	case strings.HasSuffix(entry.Filename, ".tar.gz"):
		distMediaType = MediaTypePythonDistributionSource
	case strings.HasSuffix(entry.Filename, ".egg"):
		distMediaType = MediaTypePythonDistributionEgg
	}

	dist, err := transferBlob(ctx, entry.URL, entry.Digest, distMediaType, client, target)
	if err != nil {
		return ocispec.Descriptor{}, nil, err
	}

	return dist, metadata, nil
}

// transferBlob downloads a single file and verifies the digest and then uploads the blob to OCI returning the descriptor.
func transferBlob(ctx context.Context,
	downloadURL string,
	expectedDigest digest.Digest,
	mediaType string,
	client remote.Client,
	target interface {
		content.Resolver
		content.Pusher
	},
) (ocispec.Descriptor, error) {
	log := logger.FromContext(ctx)

	// possibly short circuit if the repository already has this blob
	if expectedDigest != "" {
		// Need Blobs().Resolve() here since we need the size (.Exists() does not give us that)
		// Cannot use repository.Resolve() since that is only for manifests
		desc, err := target.Resolve(ctx, expectedDigest.String())
		if err == nil {
			desc.MediaType = mediaType
			// NOTE Resolve() populates the size from the HEAD request to the registry (we do not get that from PyPI)
			return desc, nil
		}
	}

	// fallback to downloading the file

	// TODO expectedDigest could be empty and then .Algorithm() will panic()
	alg := expectedDigest.Algorithm()
	digesters := make(map[digest.Algorithm]digest.Digester, 2)
	if alg.Available() {
		digesters[alg] = alg.Digester()
	} else {
		// Should this be fatal?  It just means we cannot validate the integrity of the file at this point.  I presume pip will validate it before installing.
		log.InfoContext(ctx, "Unsupported digest algorithm used by package index.  Unable to verify the integrity at sync-time.", "algorithm", alg)
	}

	// make sure we have the canonical digest (e.g., SHA256 for now)
	if _, exists := digesters[digest.Canonical]; !exists {
		digesters[digest.Canonical] = digest.Canonical.Digester()
	}

	// chunked upload is not supported by oras when you do not know the descriptor (digest)
	// https://github.com/oras-project/oras-go/issues/338
	// This is probably for the better so we can better validate the data before sending it to the registry.
	// This also allows us to extract metadata from the file if desired.
	// Save the package contents to a temp directory

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("building request for python file: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("downloading python file: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return ocispec.Descriptor{}, fmt.Errorf("unexpected status code: %s", res.Status)
	}

	file, err := os.CreateTemp("", "")
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("creating temporary file for python file: %w", err)
	}
	defer file.Close()
	defer os.Remove(file.Name()) // cleanup

	// Write to both the file and digesters
	writers := make([]io.Writer, 0, 1+len(digesters))
	writers = append(writers, file)
	for _, digester := range digesters {
		writers = append(writers, digester.Hash())
	}
	dest := io.MultiWriter(writers...)

	log.InfoContext(ctx, "starting download")
	_, err = io.Copy(dest, res.Body)
	log.InfoContext(ctx, "finished download")
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("writing python file to disk: %w", err)
	}

	if digester, ok := digesters[alg]; ok {
		fileDigest := digester.Digest() // this is the actual digest (same algorithm as expectedDigest)
		entryDigest := expectedDigest
		if entryDigest != fileDigest {
			return ocispec.Descriptor{}, fmt.Errorf("package index claimed file has digest %s but that does not match the downloaded file's digest of %s", entryDigest, fileDigest)
		}
	}

	info, err := file.Stat()
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	// construct the descriptor
	desc := ocispec.Descriptor{
		MediaType: mediaType,
		Digest:    digesters[digest.Canonical].Digest(),
		Size:      info.Size(),
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("failed to reset file to beginning: %w", err)
	}

	log.InfoContext(ctx, "starting upload")
	if err := target.Push(ctx, desc, file); err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("error writing blob: %w", err)
	}
	log.InfoContext(ctx, "upload finished", "digest", desc.Digest, "size", desc.Size)

	return desc, nil
}
