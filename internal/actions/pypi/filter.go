package pypi

import (
	"log/slog"
	"slices"
	"strings"

	"git.act3-ace.com/ace/data/schema/pkg/selectors"
	"git.act3-ace.com/ace/data/tool/internal/python"
	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// NOTE consider adding dependency checking (io.MultiWriter and unzip then look at the deps and add them to list to download)
// We can use https://github.com/krolaw/zipstream to read the zip with a io.Reader

// filter distribution entries (for a single project) based on the specific requirement provided by the user.
//
//nolint:sloglint
func filterDistributionEntries(log *slog.Logger,
	entries []python.DistributionEntry,
	sels selectors.LabelSelectorSet,
	allowYanked bool,
	reqs []python.Requirement,
) ([]python.DistributionEntry, error) {

	filteredPackages := make([]python.DistributionEntry, 0, len(entries))
	for _, entry := range entries {
		log := logger.V(log.With("filename", entry.Filename), 2)

		if !allowYanked && entry.Yanked != nil {
			log.Info("Skipping yanked distribution")
			continue
		}

		dist, err := python.NewDistribution(entry.Filename)
		if err != nil {
			log.Error("Skipping invalid distribution entry", "error", err.Error())
			continue
		}
		if dist == nil {
			log.Info("Ignoring unknown distribution file")
			continue
		}

		// check if the distribution satisfies any of the Requirements
		var reqSatisfied bool
		for _, requirement := range reqs {
			log := log.With("requirement", requirement)
			if python.DistributionSatisfiesRequirement(entry, dist, requirement) {
				reqSatisfied = true
				log.Info("Distribution does satisfy requirement")
			} else {
				log.Info("Distribution does not satisfy requirement")
			}
		}
		if !reqSatisfied {
			continue
		}

		// apply selectors to limit python, abi, and platform information
		labelSets := dist.Labels()
		log.Info("Computed labels", "labels", labelSets)

		ll := selectors.LabelsFromSets(labelSets)
		if !sels.MatchAny(ll) {
			log.Info("Labels do not match, skipping")
			continue
		}

		log.Info("Found matching distribution")
		filteredPackages = append(filteredPackages, entry)
	}

	// Sort them for consistency
	slices.SortStableFunc(filteredPackages, func(a, b python.DistributionEntry) int {
		return strings.Compare(a.Filename, b.Filename)
	})

	return filteredPackages, nil
}
