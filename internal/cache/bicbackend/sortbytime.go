package bicbackend

import (
	"sort"
	"time"

	"github.com/opencontainers/go-digest"
)

// CandidateWithTime is the input to BICReplacementCandidate prioritization.
type CandidateWithTime struct {
	Candidate BICReplacementCandidate // The replacement candidate
	LastSeen  time.Time               // Time the candidate was last known to exist (either read or written)
}

// candidateSortState is a local state implementing sort.Interface on candidates to prioritize,
// along with the specially-treated digest values for the implementation of sort.Interface.Less.
type candidateSortState struct {
	cs                  []CandidateWithTime // The entries to sort
	primaryDigest       digest.Digest       // The digest the user actually asked for
	untransformedDigest digest.Digest       // The untransformed digest corresponding to primaryDigest. May be "", or even equal to primaryDigest
}

func (css *candidateSortState) Len() int {
	return len(css.cs)
}

func (css *candidateSortState) Less(i, j int) bool {
	xi := css.cs[i]
	xj := css.cs[j]

	// primaryDigest entries come first, more recent first.
	// untransformedDigest entries, if untransformedDigest is set and != primaryDigest, come last, more recent entry first.
	// Other digest values are primarily sorted by time (more recent first), secondarily by digest (to provide a deterministic order)

	// First, deal with the primaryDigest/untransformedDigest cases:
	if xi.Candidate.Digest != xj.Candidate.Digest {
		// - The two digests are different, and one (or both) of the digests is primaryDigest or untransformedDigest: time does not matter
		if xi.Candidate.Digest == css.primaryDigest {
			return true
		}
		if xj.Candidate.Digest == css.primaryDigest {
			return false
		}
		if css.untransformedDigest != "" {
			if xi.Candidate.Digest == css.untransformedDigest {
				return false
			}
			if xj.Candidate.Digest == css.untransformedDigest {
				return true
			}
		}
	} else if xi.Candidate.Digest == css.primaryDigest || (css.untransformedDigest != "" && xi.Candidate.Digest == css.untransformedDigest) {
		// The two digests are the same, and are either primaryDigest or untransformedDigest: order by time
		return xi.LastSeen.After(xj.LastSeen)
	}

	// Neither of the digests are primaryDigest/untransformedDigest:
	if !xi.LastSeen.Equal(xj.LastSeen) { // Order primarily by time
		return xi.LastSeen.After(xj.LastSeen)
	}
	// Fall back to digest, if timestamps end up _exactly_ the same (how?!)
	return xi.Candidate.Digest < xj.Candidate.Digest
}

func (css *candidateSortState) Swap(i, j int) {
	css.cs[i], css.cs[j] = css.cs[j], css.cs[i]
}

// SortReplacementCandidates consumes AND DESTROYS an array of possible replacement candidates with their last known existence times,
// the primary digest the user actually asked for, and the corresponding untransformed digest (if known, possibly equal to the primary digest),
// and returns an appropriately prioritized and/or trimmed result suitable for a return value from types.BlobInfoCache.CandidateLocations.
func SortReplacementCandidates(cs []CandidateWithTime, primaryDigest, untransformedDigest digest.Digest) []BICReplacementCandidate {
	sort.Sort(&candidateSortState{
		cs:                  cs,
		primaryDigest:       primaryDigest,
		untransformedDigest: untransformedDigest,
	})

	resLength := len(cs)
	res := make([]BICReplacementCandidate, resLength)
	for i := range res {
		res[i] = cs[i].Candidate
	}
	return res
}
