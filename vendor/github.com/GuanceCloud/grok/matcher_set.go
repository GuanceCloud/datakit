package grok

import (
	"fmt"
	"math/bits"
	"sort"
	"strings"
)

// MatcherSetPattern binds a stable caller-supplied ID to a compiled matcher.
type MatcherSetPattern struct {
	ID      string
	Matcher *GrokRegexp
}

// MatcherSet holds multiple compiled matchers and a shared prefilter index.
type MatcherSet struct {
	entries         []matcherSetEntry
	linearMatchers  []*GrokRegexp
	maxMatchCount   int
	exact           map[string][]int
	anchoredBuckets map[byte][]matcherSetBucket
	atomScanner     *atomScanner
	atomPostings    []matcherSetAtomPosting
	fallback        []int
	preferLinear    bool
}

type matcherSetEntry struct {
	id        string
	matcher   *GrokRegexp
	prefilter *regexpPrefilter
	filter    matcherSetFilter
}

type matcherSetBucket struct {
	key     string
	indexes []int
}

type matcherSetAtomPosting struct {
	key     string
	indexes []int
}

// NewMatcherSet builds a shared candidate index from compiled matchers.
func NewMatcherSet(patterns []MatcherSetPattern) (*MatcherSet, error) {
	ms := &MatcherSet{
		entries:         make([]matcherSetEntry, len(patterns)),
		linearMatchers:  make([]*GrokRegexp, len(patterns)),
		exact:           make(map[string][]int),
		anchoredBuckets: make(map[byte][]matcherSetBucket),
		preferLinear:    len(patterns) <= 4,
	}

	seenIDs := make(map[string]struct{}, len(patterns))
	anchoredGroups := make(map[byte]map[string][]int)
	atomGroups := make(map[byte]map[string][]int)

	for i, pattern := range patterns {
		if pattern.ID == "" {
			return nil, fmt.Errorf("matcher set pattern %d has empty ID", i)
		}
		if pattern.Matcher == nil {
			return nil, fmt.Errorf("matcher set pattern %q is nil", pattern.ID)
		}
		if _, exists := seenIDs[pattern.ID]; exists {
			return nil, fmt.Errorf("matcher set pattern %q is duplicated", pattern.ID)
		}
		seenIDs[pattern.ID] = struct{}{}

		entry := matcherSetEntry{
			id:        pattern.ID,
			matcher:   pattern.Matcher,
			prefilter: pattern.Matcher.prefilter,
		}
		ms.entries[i] = entry
		ms.linearMatchers[i] = pattern.Matcher
		if count := pattern.Matcher.matchCount(); count > ms.maxMatchCount {
			ms.maxMatchCount = count
		}

		indexed := false
		pf := entry.prefilter
		if pf != nil {
			if pf.literalExact {
				for _, lit := range pf.literalSet {
					ms.exact[lit] = append(ms.exact[lit], i)
					indexed = true
				}
			}

			if pf.anchoredPrefix != "" {
				appendMatcherSetGroup(anchoredGroups, pf.anchoredPrefix, i)
				indexed = true
			}

			for _, atom := range matcherSetAtoms(pf) {
				appendMatcherSetGroup(atomGroups, atom, i)
				indexed = true
			}
		}

		if !indexed {
			ms.fallback = append(ms.fallback, i)
		}
	}

	ms.anchoredBuckets = flattenMatcherSetGroups(anchoredGroups)
	ms.atomPostings = flattenMatcherSetAtomGroups(atomGroups)
	ms.atomScanner = newMatcherSetAtomScanner(ms.atomPostings)
	atomIDs := matcherSetAtomIndex(ms.atomPostings)
	for i := range ms.entries {
		ms.entries[i].filter = compileMatcherSetFilter(ms.entries[i].prefilter, atomIDs)
	}
	if len(ms.exact) == 0 {
		ms.exact = nil
	}
	if len(ms.anchoredBuckets) == 0 {
		ms.anchoredBuckets = nil
	}
	if len(ms.atomPostings) == 0 {
		ms.atomPostings = nil
		ms.atomScanner = nil
	}
	if shouldPreferLinearMatcherSet(ms) {
		ms.preferLinear = true
	}

	return ms, nil
}

// MatchCount returns the largest named-capture count across all matchers in the
// set. Callers can use it to size reusable buffers passed to RunFirstTo.
func (ms *MatcherSet) MatchCount() int {
	if ms == nil {
		return 0
	}
	return ms.maxMatchCount
}

// CandidateIDs returns the in-order matcher IDs that survive the shared prefilter.
func (ms *MatcherSet) CandidateIDs(content string) []string {
	indexes := ms.candidateIndexes(content)
	out := make([]string, len(indexes))
	for i, idx := range indexes {
		out[i] = ms.entries[idx].id
	}
	return out
}

// RunFirst returns the first in-order matcher that fully matches the content.
func (ms *MatcherSet) RunFirst(content string, trimSpace bool) (string, []string, error) {
	idx, ret, err := ms.runFirstIndexTo(content, trimSpace, nil)
	if err != nil {
		return "", nil, err
	}
	return ms.entries[idx].id, ret, nil
}

// RunFirstTo returns the first in-order matcher that fully matches the content,
// reusing dst when its capacity is large enough for the matched pattern.
func (ms *MatcherSet) RunFirstTo(content string, trimSpace bool, dst []string) (string, []string, error) {
	idx, ret, err := ms.runFirstIndexTo(content, trimSpace, dst)
	if err != nil {
		return "", nil, err
	}
	return ms.entries[idx].id, ret, nil
}

func (ms *MatcherSet) runFirstIndex(content string, trimSpace bool) (int, []string, error) {
	return ms.runFirstIndexTo(content, trimSpace, nil)
}

func (ms *MatcherSet) runFirstIndexTo(content string, trimSpace bool, dst []string) (int, []string, error) {
	if len(ms.entries) == 0 {
		return -1, nil, ErrMismatch
	}
	if ms.preferLinear {
		return ms.runFirstLinearTo(content, trimSpace, dst)
	}
	if len(ms.entries) <= 64 {
		return ms.runFirstIndexBitsTo(content, trimSpace, dst)
	}

	ctx := ms.newEvalContext(content)
	seen := ms.newSeenSet()
	ms.markCandidateSet(seen, ctx)
	if ms.shouldFallBackToLinear(seen) {
		return ms.runFirstLinearTo(content, trimSpace, dst)
	}
	for i, entry := range ms.entries {
		if !seen[i] {
			continue
		}
		if !entry.filter.Accepts(ctx) {
			continue
		}
		ret, err := entry.matcher.runTo(content, trimSpace, dst)
		if err == nil {
			return i, ret, nil
		}
		if err == ErrMismatch {
			continue
		}
		return -1, nil, err
	}
	return -1, nil, ErrMismatch
}

func (ms *MatcherSet) runFirstIndexBitsTo(content string, trimSpace bool, dst []string) (int, []string, error) {
	ctx := ms.newEvalContext(content)
	seen := ms.candidateBits(ctx)
	if ms.shouldFallBackToLinearBits(seen) {
		return ms.runFirstLinearTo(content, trimSpace, dst)
	}
	for i, entry := range ms.entries {
		if seen&(uint64(1)<<uint(i)) == 0 {
			continue
		}
		if !entry.filter.Accepts(ctx) {
			continue
		}
		ret, err := entry.matcher.runTo(content, trimSpace, dst)
		if err == nil {
			return i, ret, nil
		}
		if err == ErrMismatch {
			continue
		}
		return -1, nil, err
	}
	return -1, nil, ErrMismatch
}

func (ms *MatcherSet) runFirstLinear(content string, trimSpace bool) (int, []string, error) {
	return ms.runFirstLinearTo(content, trimSpace, nil)
}

func (ms *MatcherSet) runFirstLinearTo(content string, trimSpace bool, dst []string) (int, []string, error) {
	for i, matcher := range ms.linearMatchers {
		ret, err := matcher.runTo(content, trimSpace, dst)
		if err == nil {
			return i, ret, nil
		}
		if err == ErrMismatch {
			continue
		}
		return -1, nil, err
	}
	return -1, nil, ErrMismatch
}

func (ms *MatcherSet) candidateIndexes(content string) []int {
	if len(ms.entries) == 0 {
		return nil
	}

	ctx := ms.newEvalContext(content)
	seen := ms.newSeenSet()
	ms.markCandidateSet(seen, ctx)
	if ms.shouldFallBackToLinear(seen) {
		return ms.linearCandidateIndexes(ctx)
	}

	out := make([]int, 0, len(ms.entries))
	for i, entry := range ms.entries {
		if !seen[i] {
			continue
		}
		if !entry.filter.Accepts(ctx) {
			continue
		}
		out = append(out, i)
	}
	return out
}

func (ms *MatcherSet) linearCandidateIndexes(ctx matcherSetEvalContext) []int {
	out := make([]int, 0, len(ms.entries))
	for i, entry := range ms.entries {
		if !entry.filter.Accepts(ctx) {
			continue
		}
		out = append(out, i)
	}
	return out
}

func (ms *MatcherSet) newSeenSet() []bool {
	if len(ms.entries) == 0 {
		return nil
	}
	return make([]bool, len(ms.entries))
}

func (ms *MatcherSet) newEvalContext(content string) matcherSetEvalContext {
	ctx := matcherSetEvalContext{Content: content}
	if ms.atomScanner != nil && len(ms.atomPostings) > 0 {
		if len(ms.atomPostings) <= 64 {
			ctx.UseBits = true
			ctx.AtomBits = ms.atomScanner.ScanBits(content)
		} else {
			ctx.AtomHits = make([]bool, len(ms.atomPostings))
			ms.atomScanner.ScanInto(content, ctx.AtomHits)
		}
	}
	return ctx
}

func (ms *MatcherSet) markCandidateSet(seen []bool, ctx matcherSetEvalContext) {
	if len(seen) == 0 {
		return
	}

	if len(ms.exact) > 0 {
		markMatcherSetIndexes(seen, ms.exact[ctx.Content])
	}

	if len(ctx.Content) > 0 && len(ms.anchoredBuckets) > 0 {
		for _, bucket := range ms.anchoredBuckets[ctx.Content[0]] {
			if strings.HasPrefix(ctx.Content, bucket.key) {
				markMatcherSetIndexes(seen, bucket.indexes)
			}
		}
	}

	if ctx.UseBits {
		for i := 0; i < len(ms.atomPostings) && i < 64; i++ {
			if ctx.AtomBits&(uint64(1)<<i) == 0 {
				continue
			}
			markMatcherSetIndexes(seen, ms.atomPostings[i].indexes)
		}
	} else if len(ctx.AtomHits) > 0 {
		for i, hit := range ctx.AtomHits {
			if !hit {
				continue
			}
			markMatcherSetIndexes(seen, ms.atomPostings[i].indexes)
		}
	}

	markMatcherSetIndexes(seen, ms.fallback)
}

func (ms *MatcherSet) candidateBits(ctx matcherSetEvalContext) uint64 {
	var seen uint64

	if len(ms.exact) > 0 {
		seen = markMatcherSetIndexBits(seen, ms.exact[ctx.Content])
	}

	if len(ctx.Content) > 0 && len(ms.anchoredBuckets) > 0 {
		for _, bucket := range ms.anchoredBuckets[ctx.Content[0]] {
			if strings.HasPrefix(ctx.Content, bucket.key) {
				seen = markMatcherSetIndexBits(seen, bucket.indexes)
			}
		}
	}

	if ctx.UseBits {
		for i := 0; i < len(ms.atomPostings) && i < 64; i++ {
			if ctx.AtomBits&(uint64(1)<<uint(i)) == 0 {
				continue
			}
			seen = markMatcherSetIndexBits(seen, ms.atomPostings[i].indexes)
		}
	} else if len(ctx.AtomHits) > 0 {
		for i, hit := range ctx.AtomHits {
			if !hit {
				continue
			}
			seen = markMatcherSetIndexBits(seen, ms.atomPostings[i].indexes)
		}
	}

	return markMatcherSetIndexBits(seen, ms.fallback)
}

func (ms *MatcherSet) shouldFallBackToLinear(seen []bool) bool {
	entryCount := len(ms.entries)
	if entryCount == 0 || entryCount > 16 {
		return false
	}
	candidates := 0
	for _, ok := range seen {
		if ok {
			candidates++
		}
	}
	return candidates*3 >= entryCount*2
}

func (ms *MatcherSet) shouldFallBackToLinearBits(seen uint64) bool {
	entryCount := len(ms.entries)
	if entryCount == 0 || entryCount > 16 {
		return false
	}
	return bits.OnesCount64(seen)*3 >= entryCount*2
}

func matcherSetAtoms(pf *regexpPrefilter) []string {
	if pf == nil {
		return nil
	}

	seen := make(map[string]struct{}, 1+len(pf.literalSet)+len(pf.required))
	out := make([]string, 0, 1+len(pf.literalSet)+len(pf.required))

	appendAtom := func(atom string) {
		if atom == "" {
			return
		}
		if _, exists := seen[atom]; exists {
			return
		}
		seen[atom] = struct{}{}
		out = append(out, atom)
	}

	if pf.literalPrefix != "" {
		appendAtom(pf.literalPrefix)
	}
	if !pf.literalExact {
		for _, lit := range pf.literalSet {
			appendAtom(lit)
		}
	}
	for _, lit := range pf.required {
		if lit == pf.anchoredPrefix {
			continue
		}
		appendAtom(lit)
	}

	return out
}

func appendMatcherSetGroup(groups map[byte]map[string][]int, key string, idx int) {
	if key == "" {
		return
	}
	bucketKey := key[0]
	bucket := groups[bucketKey]
	if bucket == nil {
		bucket = make(map[string][]int)
		groups[bucketKey] = bucket
	}
	bucket[key] = append(bucket[key], idx)
}

func flattenMatcherSetGroups(groups map[byte]map[string][]int) map[byte][]matcherSetBucket {
	if len(groups) == 0 {
		return nil
	}

	out := make(map[byte][]matcherSetBucket, len(groups))
	for bucketKey, group := range groups {
		buckets := make([]matcherSetBucket, 0, len(group))
		for key, indexes := range group {
			buckets = append(buckets, matcherSetBucket{key: key, indexes: indexes})
		}
		out[bucketKey] = buckets
	}
	return out
}

func flattenMatcherSetAtomGroups(groups map[byte]map[string][]int) []matcherSetAtomPosting {
	if len(groups) == 0 {
		return nil
	}

	keys := make([]string, 0)
	for _, group := range groups {
		for key := range group {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	out := make([]matcherSetAtomPosting, 0, len(keys))
	seen := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		group := groups[key[0]]
		out = append(out, matcherSetAtomPosting{
			key:     key,
			indexes: group[key],
		})
	}
	return out
}

func newMatcherSetAtomScanner(postings []matcherSetAtomPosting) *atomScanner {
	if len(postings) == 0 {
		return nil
	}
	atoms := make([]string, 0, len(postings))
	for _, posting := range postings {
		atoms = append(atoms, posting.key)
	}
	return newAtomScanner(atoms)
}

func matcherSetAtomIndex(postings []matcherSetAtomPosting) map[string]int {
	if len(postings) == 0 {
		return nil
	}
	out := make(map[string]int, len(postings))
	for i, posting := range postings {
		out[posting.key] = i
	}
	return out
}

func markMatcherSetIndexes(dst []bool, indexes []int) {
	for _, idx := range indexes {
		dst[idx] = true
	}
}

func markMatcherSetIndexBits(seen uint64, indexes []int) uint64 {
	for _, idx := range indexes {
		seen |= uint64(1) << uint(idx)
	}
	return seen
}

func shouldPreferLinearMatcherSet(ms *MatcherSet) bool {
	if ms == nil {
		return true
	}
	entryCount := len(ms.entries)
	if entryCount <= 4 {
		return true
	}
	exactBuckets := len(ms.exact)
	anchoredBuckets := matcherSetBucketCount(ms.anchoredBuckets)
	if exactBuckets == 0 && anchoredBuckets == 0 && entryCount <= 8 {
		return true
	}
	indexedEntries := entryCount - len(ms.fallback)
	bestBucket := matcherSetBestBucketSize(ms)
	if bestBucket == 0 {
		return true
	}
	if entryCount <= 12 {
		if bestBucket*3 >= entryCount*2 {
			return true
		}
		if indexedEntries*4 < entryCount*3 {
			return true
		}
	}
	selectiveAtoms := matcherSetSelectivePostingCount(ms.atomPostings, entryCount)
	score := exactBuckets*4 + anchoredBuckets*3 + selectiveAtoms*2 + indexedEntries - len(ms.fallback)*3 - bestBucket
	return score < entryCount*2
}

func matcherSetBucketCount(groups map[byte][]matcherSetBucket) int {
	total := 0
	for _, buckets := range groups {
		total += len(buckets)
	}
	return total
}

func matcherSetSelectivePostingCount(postings []matcherSetAtomPosting, totalEntries int) int {
	if totalEntries <= 0 {
		return 0
	}
	total := 0
	for _, posting := range postings {
		if len(posting.indexes)*2 <= totalEntries {
			total++
		}
	}
	return total
}

func matcherSetBestBucketSize(ms *MatcherSet) int {
	if ms == nil {
		return 0
	}

	best := 0
	update := func(size int) {
		if size <= 0 {
			return
		}
		if best == 0 || size < best {
			best = size
		}
	}

	for _, indexes := range ms.exact {
		update(len(indexes))
	}
	for _, buckets := range ms.anchoredBuckets {
		for _, bucket := range buckets {
			update(len(bucket.indexes))
		}
	}
	for _, posting := range ms.atomPostings {
		update(len(posting.indexes))
	}
	return best
}
