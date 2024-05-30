package pypi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/retry"

	"gitlab.com/act3-ai/asce/data/tool/internal/python"
	"gitlab.com/act3-ai/asce/data/tool/internal/ui"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// ToPyPI represents the pypi to-pypi action.
type ToPyPI struct {
	*Action

	// DryRun prevents uploading distribution files to the package index
	DryRun bool

	client remote.Client
}

// Run the action.
func (action *ToPyPI) Run(ctx context.Context, ociRepo, pypi string) error {
	// log := logger.FromContext(ctx)

	action.client = retry.DefaultClient

	src, err := action.Config.Repository(ctx, ociRepo)
	if err != nil {
		return err
	}

	// list the tags
	projects, err := registry.Tags(ctx, src)
	if err != nil {
		return fmt.Errorf("listing projects from tags: %w", err)
	}

	// HACK: Artifactory returns non-tags from the tag listing API so we need to validate them to remove bogus project names
	projects = slices.DeleteFunc(projects, func(project string) bool {
		// for now we just filter out tags with a ":" (tags are not allowed to have a colon)
		// see https://docs.docker.com/engine/reference/commandline/tag/
		return strings.Contains(project, ":")
	})

	for _, project := range projects {
		// TODO filter projects by name

		// process the project denoted by tag
		if err := action.processProject(ctx, src, pypi, project); err != nil {
			return err
		}
	}

	return nil
}

func (action *ToPyPI) processProject(ctx context.Context, src oras.ReadOnlyTarget, pypi string, project string) error {
	log := logger.FromContext(ctx)

	task := ui.FromContextOrNoop(ctx).SubTask(project)
	defer task.Complete()

	task.Info("Processing")

	pyDistIdx, err := newPythonDistributionIndex(ctx, src, false, project)
	if err != nil {
		return err
	}

	distributions, err := pyDistIdx.GetDistributions(action.AllowYanked)
	if err != nil {
		return err
	}
	// TODO filter distributions by label selectors (similar to how it is done in to-oci)

	log.InfoContext(ctx, "Found distributions from OCI", "count", len(distributions))

	// find existing packages
	existing := map[string]struct{}{}

	{
		// find the package on the package index
		existingDists, err := python.RetrieveDistributions(ctx, action.client, pypi+"/simple", project)
		switch {
		case errors.Is(err, python.ErrProjectNotFound):
			// missing projects have not packages
		case err != nil:
			return err
		default: // err == nil
			for _, d := range existingDists {
				existing[d.Filename] = struct{}{}
			}
		}
	}

	// mutates the distributions slice
	missingDistributions := slices.DeleteFunc(distributions, func(de python.DistributionEntry) bool {
		_, ok := existing[de.Filename]
		return ok
	})

	task.Infof("Found %d packages in OCI and %d in PyPI (%d will be uploaded)", len(distributions), len(existing), len(missingDistributions))

	for _, entry := range missingDistributions {
		if err := action.uploadPythonEntry(ctx, task, pyDistIdx, entry, pypi); err != nil {
			return err
		}
	}

	return nil
}

func (action *ToPyPI) uploadPythonEntry(ctx context.Context, task *ui.Task, pyDistIdx *pythonDistributionIndex, entry python.DistributionEntry, pypi string) error {
	thisTask := task.SubTask(entry.Filename)
	defer thisTask.Complete()

	man, ok := pyDistIdx.LookupDistribution(entry.Filename)
	if !ok {
		// FIXME this function is called with the results of GetDistributions()
		// Due to the sync version, we might have values in GetDistributions that is different than what we see in the lookup (which only has the latest sync version)
		panic(fmt.Sprintf("distribution %q missing", entry.Filename))
	}

	desc, _, err := pyDistIdx.FetchDistribution(ctx, man)
	if err != nil {
		return err
	}

	// TODO we could apply the same filters that we do for "ace-dt pypi to-oci" here to restrict which packages we upload

	getContent := func() (io.ReadCloser, error) {
		rc, err := pyDistIdx.target.Fetch(ctx, desc)
		if err != nil {
			return nil, fmt.Errorf("fetching distribution information from OCI: %w", err)
		}

		return rc, nil
	}

	if action.DryRun {
		thisTask.Info("Would upload")
		return nil
	}

	thisTask.Info("Uploading")
	return python.Upload(ctx, action.client, getContent, pypi, entry)
}
