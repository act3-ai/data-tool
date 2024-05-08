package bicbackend

import (
	"context"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	"git.act3-ace.com/ace/go-common/pkg/test"
)

const (
	digestUnknown              = digest.Digest("sha256:1111111111111111111111111111111111111111111111111111111111111111")
	digestUntransformed        = digest.Digest("sha256:2222222222222222222222222222222222222222222222222222222222222222")
	digestTransformedA         = digest.Digest("sha256:3333333333333333333333333333333333333333333333333333333333333333")
	digestTransformedB         = digest.Digest("sha256:4444444444444444444444444444444444444444444444444444444444444444")
	digestTransformedUnrelated = digest.Digest("sha256:5555555555555555555555555555555555555555555555555555555555555555")
	transformerNameU           = "transformerName/U"
	transformerNameA           = "transformerName/A"
	transformerNameB           = "transformerName/B"
	transformerNameCU          = "transformerName/CU"
)

// testGenericCache runs an implementation-independent set of tests, given a
// newTestCache, which can be called repeatedly and always returns a fresh cache instance.
func testGenericCache(t *testing.T, newTestCache func(t *testing.T) BlobInfoCache) {
	t.Helper()
	subs := []struct {
		name string
		fn   func(t *testing.T, cache BlobInfoCache)
	}{
		{"UntransformedDigest", testGenericUntransformedDigest},
		{"RecordDigestUntransformedPair", testGenericRecordDigestUntransformedPair},
		{"RecordKnownLocations", testGenericRecordKnownLocations},
		{"CandidateLocations", testGenericCandidateLocations},
		{"CandidateLocations2", testGenericCandidateLocations2},
	}

	for _, s := range subs {
		t.Run(s.name, func(t *testing.T) {
			cache := newTestCache(t)
			cache.Open()
			defer cache.Close()
			s.fn(t, cache)
		})
	}
}

func testGenericUntransformedDigest(t *testing.T, cache BlobInfoCache) {
	t.Helper()
	ctx := logger.NewContext(context.Background(), test.Logger(t, 0))
	// Nothing is known.
	assert.Equal(t, digest.Digest(""), cache.UntransformedDigest(ctx, digestUnknown))

	cache.RecordDigestUntransformedPair(ctx, digestTransformedA, digestUntransformed)
	cache.RecordDigestUntransformedPair(ctx, digestTransformedB, digestUntransformed)
	// Known transformed→untransformed mapping
	assert.Equal(t, digestUntransformed, cache.UntransformedDigest(ctx, digestTransformedA))
	assert.Equal(t, digestUntransformed, cache.UntransformedDigest(ctx, digestTransformedB))
	// This implicitly marks digestUntransformed as untransformed.
	assert.Equal(t, digestUntransformed, cache.UntransformedDigest(ctx, digestUntransformed))

	// Known untransformed→self mapping
	cache.RecordDigestUntransformedPair(ctx, digestTransformedUnrelated, digestTransformedUnrelated)
	assert.Equal(t, digestTransformedUnrelated, cache.UntransformedDigest(ctx, digestTransformedUnrelated))
}

func testGenericRecordDigestUntransformedPair(t *testing.T, cache BlobInfoCache) {
	t.Helper()
	ctx := logger.NewContext(context.Background(), test.Logger(t, 0))
	for i := 0; i < 2; i++ { // Record the same data twice to ensure redundant writes don’t break things.
		// Known transformed→untransformed mapping
		cache.RecordDigestUntransformedPair(ctx, digestTransformedA, digestUntransformed)
		assert.Equal(t, digestUntransformed, cache.UntransformedDigest(ctx, digestTransformedA))
		// Two mappings to the same untransformed digest
		cache.RecordDigestUntransformedPair(ctx, digestTransformedB, digestUntransformed)
		assert.Equal(t, digestUntransformed, cache.UntransformedDigest(ctx, digestTransformedB))

		// Mapping an untransformed digest to self
		cache.RecordDigestUntransformedPair(ctx, digestUntransformed, digestUntransformed)
		assert.Equal(t, digestUntransformed, cache.UntransformedDigest(ctx, digestUntransformed))
	}
}

func testGenericRecordKnownLocations(t *testing.T, cache BlobInfoCache) {
	t.Helper()
	ctx := logger.NewContext(context.Background(), test.Logger(t, 0))
	transport := "==BlobInfoCache TestTransport"
	for i := 0; i < 2; i++ { // Record the same data twice to ensure redundant writes don’t break things.
		for _, scopeName := range []string{"A", "B"} { // Run the test in two different scopes to verify they don't affect each other.
			scope := BICContentScope{Opaque: scopeName}
			for _, digest := range []digest.Digest{digestTransformedA, digestTransformedB} { // Two different digests should not affect each other either.
				lr1 := BICLocationReference{Opaque: scopeName + "1"}
				lr2 := BICLocationReference{Opaque: scopeName + "2"}
				cache.RecordKnownLocation(ctx, transport, scope, digest, lr2)
				cache.RecordKnownLocation(ctx, transport, scope, digest, lr1)
				assert.Equal(t, []BICReplacementCandidate{
					{Digest: digest, Location: lr1, TransformerName: UnknownTransformer},
					{Digest: digest, Location: lr2, TransformerName: UnknownTransformer},
				}, cache.CandidateLocations(ctx, transport, scope, digest, false))
			}
		}
	}
}

// candidate is a shorthand for BICReplacementCandidate.
type candidate struct {
	d  digest.Digest
	tn string
	lr string
}

func assertCandidatesMatch(t *testing.T, scopeName string, expected []candidate, actual []BICReplacementCandidate) {
	t.Helper()
	e := make([]BICReplacementCandidate, len(expected))
	for i, ev := range expected {
		e[i] = BICReplacementCandidate{Digest: ev.d, TransformerName: ev.tn, Location: BICLocationReference{Opaque: scopeName + ev.lr}}
	}
	assert.Equal(t, e, actual)
}

func testGenericCandidateLocations(t *testing.T, cache BlobInfoCache) {
	t.Helper()
	ctx := logger.NewContext(context.Background(), test.Logger(t, 0))
	transport := "==BlobInfoCache TestTransport"
	cache.RecordDigestUntransformedPair(ctx, digestTransformedA, digestUntransformed)
	cache.RecordDigestUntransformedPair(ctx, digestTransformedB, digestUntransformed)
	cache.RecordDigestUntransformedPair(ctx, digestUntransformed, digestUntransformed)
	digestNameSet := []struct {
		n string
		d digest.Digest
	}{
		{"U", digestUntransformed},
		{"A", digestTransformedA},
		{"B", digestTransformedB},
		{"CU", digestTransformedUnrelated},
	}

	for _, scopeName := range []string{"A", "B"} { // Run the test in two different scopes to verify they don't affect each other.
		scope := BICContentScope{Opaque: scopeName}
		// Nothing is known.
		assert.Equal(t, []BICReplacementCandidate{}, cache.CandidateLocations(ctx, transport, scope, digestUnknown, false))
		assert.Equal(t, []BICReplacementCandidate{}, cache.CandidateLocations(ctx, transport, scope, digestUnknown, true))

		// Record "2" entries before "1" entries; then results should sort "1" (more recent) before "2" (older)
		for _, suffix := range []string{"2", "1"} {
			for _, e := range digestNameSet {
				cache.RecordKnownLocation(ctx, transport, scope, e.d, BICLocationReference{Opaque: scopeName + e.n + suffix})
			}
		}

		// No substitutions allowed:
		for _, e := range digestNameSet {
			assertCandidatesMatch(t, scopeName, []candidate{
				{d: e.d, tn: UnknownTransformer, lr: e.n + "1"}, {d: e.d, tn: UnknownTransformer, lr: e.n + "2"},
			}, cache.CandidateLocations(ctx, transport, scope, e.d, false))
		}

		// With substitutions: The original digest is always preferred, then other transformed, then the untransformed one.
		assertCandidatesMatch(t, scopeName, []candidate{
			{d: digestTransformedA, tn: UnknownTransformer, lr: "A1"}, {d: digestTransformedA, tn: UnknownTransformer, lr: "A2"},
			{d: digestTransformedB, tn: UnknownTransformer, lr: "B1"}, {d: digestTransformedB, tn: UnknownTransformer, lr: "B2"},
			{d: digestUntransformed, tn: UnknownTransformer, lr: "U1"}, {d: digestUntransformed, tn: UnknownTransformer, lr: "U2"},
		}, cache.CandidateLocations(ctx, transport, scope, digestTransformedA, true))

		assertCandidatesMatch(t, scopeName, []candidate{
			{d: digestTransformedB, tn: UnknownTransformer, lr: "B1"}, {d: digestTransformedB, tn: UnknownTransformer, lr: "B2"},
			{d: digestTransformedA, tn: UnknownTransformer, lr: "A1"}, {d: digestTransformedA, tn: UnknownTransformer, lr: "A2"},
			{d: digestUntransformed, tn: UnknownTransformer, lr: "U1"}, {d: digestUntransformed, tn: UnknownTransformer, lr: "U2"},
		}, cache.CandidateLocations(ctx, transport, scope, digestTransformedB, true))

		assertCandidatesMatch(t, scopeName, []candidate{
			{d: digestUntransformed, tn: UnknownTransformer, lr: "U1"}, {d: digestUntransformed, tn: UnknownTransformer, lr: "U2"},
			// "1" entries were added after "2", and A/Bs are sorted in the reverse of digestNameSet order
			{d: digestTransformedB, tn: UnknownTransformer, lr: "B1"},
			{d: digestTransformedA, tn: UnknownTransformer, lr: "A1"},
			{d: digestTransformedB, tn: UnknownTransformer, lr: "B2"},
			{d: digestTransformedA, tn: UnknownTransformer, lr: "A2"},
		}, cache.CandidateLocations(ctx, transport, scope, digestUntransformed, true))

		// Locations are known, but no relationships
		assertCandidatesMatch(t, scopeName, []candidate{
			{d: digestTransformedUnrelated, tn: UnknownTransformer, lr: "CU1"},
			{d: digestTransformedUnrelated, tn: UnknownTransformer, lr: "CU2"},
		}, cache.CandidateLocations(ctx, transport, scope, digestTransformedUnrelated, true))
	}
}

func testGenericCandidateLocations2(t *testing.T, cache BlobInfoCache) {
	t.Helper()
	ctx := logger.NewContext(context.Background(), test.Logger(t, 0))
	transport := "==BlobInfoCache TestTransport"
	cache.RecordDigestUntransformedPair(ctx, digestTransformedA, digestUntransformed)
	cache.RecordDigestUntransformedPair(ctx, digestTransformedB, digestUntransformed)
	cache.RecordDigestUntransformedPair(ctx, digestUntransformed, digestUntransformed)
	digestNameSet := []struct {
		n string
		d digest.Digest
		m string
	}{
		{"U", digestUntransformed, transformerNameU},
		{"A", digestTransformedA, transformerNameA},
		{"B", digestTransformedB, transformerNameB},
		{"CU", digestTransformedUnrelated, transformerNameCU},
	}

	for scopeIndex, scopeName := range []string{"A", "B", "C"} { // Run the test in two different scopes to verify they don't affect each other.
		scope := BICContentScope{Opaque: scopeName}

		// Nothing is known.
		assert.Equal(t, []BICReplacementCandidate{}, cache.CandidateLocations(ctx, transport, scope, digestUnknown, false))
		assert.Equal(t, []BICReplacementCandidate{}, cache.CandidateLocations(ctx, transport, scope, digestUnknown, true))

		// Record "2" entries before "1" entries; then results should sort "1" (more recent) before "2" (older)
		for _, suffix := range []string{"2", "1"} {
			for _, e := range digestNameSet {
				cache.RecordKnownLocation(ctx, transport, scope, e.d, BICLocationReference{Opaque: scopeName + e.n + suffix})
			}
		}

		// Clear any "known" compression values, except on the first loop where they've never been set.
		// This probably triggers “Compressor for blob with digest … previously recorded as …, now unknown” warnings here, for test purposes;
		// that shouldn’t happen in real-world usage.
		if scopeIndex != 0 {
			for _, e := range digestNameSet {
				cache.RecordDigestTransformerName(ctx, e.d, UnknownTransformer)
			}
		}

		// No substitutions allowed:
		for _, e := range digestNameSet {
			assertCandidatesMatch(t, scopeName, []candidate{
				{d: e.d, tn: UnknownTransformer, lr: e.n + "1"},
				{d: e.d, tn: UnknownTransformer, lr: e.n + "2"},
			}, cache.CandidateLocations(ctx, transport, scope, e.d, false))
		}

		// With substitutions: The original digest is always preferred, then other transformed, then the untransformed one.
		assertCandidatesMatch(t, scopeName, []candidate{
			{d: digestTransformedA, tn: UnknownTransformer, lr: "A1"},
			{d: digestTransformedA, tn: UnknownTransformer, lr: "A2"},
			{d: digestTransformedB, tn: UnknownTransformer, lr: "B1"},
			{d: digestTransformedB, tn: UnknownTransformer, lr: "B2"},
			{d: digestUntransformed, tn: UnknownTransformer, lr: "U1"},
			{d: digestUntransformed, tn: UnknownTransformer, lr: "U2"},
		}, cache.CandidateLocations(ctx, transport, scope, digestTransformedA, true))
		// Unknown compression -> no candidates

		assertCandidatesMatch(t, scopeName, []candidate{
			{d: digestTransformedB, tn: UnknownTransformer, lr: "B1"},
			{d: digestTransformedB, tn: UnknownTransformer, lr: "B2"},
			{d: digestTransformedA, tn: UnknownTransformer, lr: "A1"},
			{d: digestTransformedA, tn: UnknownTransformer, lr: "A2"},
			{d: digestUntransformed, tn: UnknownTransformer, lr: "U1"},
			{d: digestUntransformed, tn: UnknownTransformer, lr: "U2"},
		}, cache.CandidateLocations(ctx, transport, scope, digestTransformedB, true))

		assertCandidatesMatch(t, scopeName, []candidate{
			{d: digestUntransformed, tn: UnknownTransformer, lr: "U1"},
			{d: digestUntransformed, tn: UnknownTransformer, lr: "U2"},
			// "1" entries were added after "2", and A/Bs are sorted in the reverse of digestNameSet order
			{d: digestTransformedB, tn: UnknownTransformer, lr: "B1"},
			{d: digestTransformedA, tn: UnknownTransformer, lr: "A1"},
			{d: digestTransformedB, tn: UnknownTransformer, lr: "B2"},
			{d: digestTransformedA, tn: UnknownTransformer, lr: "A2"},
		}, cache.CandidateLocations(ctx, transport, scope, digestUntransformed, true))

		// Locations are known, but no relationships
		assertCandidatesMatch(t, scopeName, []candidate{
			{d: digestTransformedUnrelated, tn: UnknownTransformer, lr: "CU1"},
			{d: digestTransformedUnrelated, tn: UnknownTransformer, lr: "CU2"},
		}, cache.CandidateLocations(ctx, transport, scope, digestTransformedUnrelated, true))

		// Set the "known" compression values
		for _, e := range digestNameSet {
			cache.RecordDigestTransformerName(ctx, e.d, e.m)
		}

		// No substitutions allowed:
		for _, e := range digestNameSet {
			assertCandidatesMatch(t, scopeName, []candidate{
				{d: e.d, tn: e.m, lr: e.n + "1"},
				{d: e.d, tn: e.m, lr: e.n + "2"},
			}, cache.CandidateLocations(ctx, transport, scope, e.d, false))
		}

		// With substitutions: The original digest is always preferred, then other transformed, then the untransformed one.
		assertCandidatesMatch(t, scopeName, []candidate{
			{d: digestTransformedA, tn: transformerNameA, lr: "A1"},
			{d: digestTransformedA, tn: transformerNameA, lr: "A2"},
			{d: digestTransformedB, tn: transformerNameB, lr: "B1"},
			{d: digestTransformedB, tn: transformerNameB, lr: "B2"},
			{d: digestUntransformed, tn: transformerNameU, lr: "U1"},
			{d: digestUntransformed, tn: transformerNameU, lr: "U2"},
		}, cache.CandidateLocations(ctx, transport, scope, digestTransformedA, true))

		assertCandidatesMatch(t, scopeName, []candidate{
			{d: digestTransformedB, tn: transformerNameB, lr: "B1"},
			{d: digestTransformedB, tn: transformerNameB, lr: "B2"},
			{d: digestTransformedA, tn: transformerNameA, lr: "A1"},
			{d: digestTransformedA, tn: transformerNameA, lr: "A2"},
			{d: digestUntransformed, tn: transformerNameU, lr: "U1"},
			{d: digestUntransformed, tn: transformerNameU, lr: "U2"},
		}, cache.CandidateLocations(ctx, transport, scope, digestTransformedB, true))

		assertCandidatesMatch(t, scopeName, []candidate{
			{d: digestUntransformed, tn: transformerNameU, lr: "U1"},
			{d: digestUntransformed, tn: transformerNameU, lr: "U2"},
			// "1" entries were added after "2", and A/Bs are sorted in the reverse of digestNameSet order
			{d: digestTransformedB, tn: transformerNameB, lr: "B1"},
			{d: digestTransformedA, tn: transformerNameA, lr: "A1"},
			{d: digestTransformedB, tn: transformerNameB, lr: "B2"},
			{d: digestTransformedA, tn: transformerNameA, lr: "A2"},
		}, cache.CandidateLocations(ctx, transport, scope, digestUntransformed, true))

		//// Locations are known, but no relationships
		assertCandidatesMatch(t, scopeName, []candidate{
			{d: digestTransformedUnrelated, tn: transformerNameCU, lr: "CU1"},
			{d: digestTransformedUnrelated, tn: transformerNameCU, lr: "CU2"},
		}, cache.CandidateLocations(ctx, transport, scope, digestTransformedUnrelated, true))
	}
}
