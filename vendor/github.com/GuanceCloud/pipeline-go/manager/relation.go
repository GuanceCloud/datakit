// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package manager for managing pipeline scripts
package manager

import (
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/GuanceCloud/cliutils/point"
)

// var remoteRelation = &PipelineRelation{}

type wildcardRelation struct {
	pattern      string
	patternRunes []rune
	prefix       string
	suffix       string
	name         string
	literalCount int
	minLen       int
	starCount    int
	patternLen   int
}

type wildcardTrieNode struct {
	children  map[rune]*wildcardTrieNode
	rules     []wildcardRelation
	indexed   bool
	exactLen  map[int][]wildcardRelation
	starLens  []int
	starByLen map[int][]wildcardRelation
}

type literalAnchorNode struct {
	children map[rune]int
	fail     int
	outputs  []int
}

type literalAnchorScratch struct {
	marks []uint32
	gen   uint32
}

type literalAnchorIndex struct {
	exact map[string]string

	generic []wildcardRelation

	nodes       []literalAnchorNode
	anchorRules [][]wildcardRelation

	scratchPool sync.Pool
}

type wildcardIndex struct {
	prefixTrie *wildcardTrieNode
	suffixTrie *wildcardTrieNode
	generic    []wildcardRelation
	anchor     *literalAnchorIndex
}

type relationState struct {
	updateAt  int64
	relation  map[point.Category]map[string]string
	wildcards map[point.Category]*wildcardIndex
	cache     *relationQueryCache
}

type relationQueryCacheKey struct {
	cat    point.Category
	source string
}

type relationQueryCacheValue struct {
	name string
	ok   bool
}

type relationQueryCache struct {
	limit uint32
	size  atomic.Uint32
	data  sync.Map
}

type ScriptRelation struct {
	// map[<category>]: map[<source|pattern>]: <name>
	relation map[point.Category]map[string]string

	// Compiled wildcard indices. Exact match is still checked first.
	wildcards map[point.Category]*wildcardIndex

	updateAt int64

	state atomic.Value

	rwMutex sync.RWMutex
}

func NewPipelineRelation() *ScriptRelation {
	return &ScriptRelation{}
}

func (relation *ScriptRelation) UpdateAt() int64 {
	if state := relation.loadState(); state != nil {
		return state.updateAt
	}

	relation.rwMutex.RLock()
	defer relation.rwMutex.RUnlock()

	return relation.updateAt
}

func (relation *ScriptRelation) UpdateRelation(updateAt int64, rel map[point.Category]map[string]string) {
	relation.rwMutex.Lock()
	defer relation.rwMutex.Unlock()

	relation.updateAt = updateAt

	// reset relation
	relation.relation = map[point.Category]map[string]string{}
	relation.wildcards = map[point.Category]*wildcardIndex{}

	for cat, relat := range rel {
		m := map[string]string{}
		relation.relation[cat] = m
		var index wildcardIndex
		for source, name := range relat {
			m[source] = name
			if hasWildcard(source) {
				index.add(newWildcardRelation(source, name))
			}
		}
		if index.hasRules() {
			index.finalize()
			relation.wildcards[cat] = &index
		}
	}

	relation.state.Store(&relationState{
		updateAt:  relation.updateAt,
		relation:  relation.relation,
		wildcards: relation.wildcards,
		cache:     newRelationQueryCache(4096),
	})
}

func (relation *ScriptRelation) Query(cat point.Category, source string) (string, bool) {
	if state := relation.loadState(); state != nil {
		return queryRelationStateWithCache(state.relation, state.wildcards, state.cache, cat, source)
	}

	relation.rwMutex.RLock()
	defer relation.rwMutex.RUnlock()

	return queryRelationState(relation.relation, relation.wildcards, cat, source)
}

func (relation *ScriptRelation) loadState() *relationState {
	state := relation.state.Load()
	if state == nil {
		return nil
	}

	if typed, ok := state.(*relationState); ok {
		return typed
	}

	return nil
}

func queryRelationState(rels map[point.Category]map[string]string, wildcards map[point.Category]*wildcardIndex, cat point.Category, source string) (string, bool) {
	return queryRelationStateWithCache(rels, wildcards, nil, cat, source)
}

func queryRelationStateWithCache(rels map[point.Category]map[string]string, wildcards map[point.Category]*wildcardIndex, cache *relationQueryCache, cat point.Category, source string) (string, bool) {
	if v, ok := rels[cat]; ok {
		if name, ok := v[source]; ok {
			return name, true
		}

		if index, ok := wildcards[cat]; ok && index != nil {
			if cache != nil {
				if cached, ok := cache.Load(cat, source); ok {
					return cached.name, cached.ok
				}
			}

			sourceRunes := []rune(source)
			name, ok := index.query(source, sourceRunes)
			if cache != nil {
				cache.Store(cat, source, relationQueryCacheValue{name: name, ok: ok})
			}
			return name, ok
		} else {
			// Keep compatibility for callers/tests that write relation.relation directly.
			sourceRunes := []rune(source)
			if name, ok := queryWildcardMap(v, source, sourceRunes); ok {
				return name, true
			}
		}
	}

	return "", false
}

func newRelationQueryCache(limit uint32) *relationQueryCache {
	if limit == 0 {
		return nil
	}

	return &relationQueryCache{limit: limit}
}

func (cache *relationQueryCache) Load(cat point.Category, source string) (relationQueryCacheValue, bool) {
	if cache == nil {
		return relationQueryCacheValue{}, false
	}

	val, ok := cache.data.Load(relationQueryCacheKey{
		cat:    cat,
		source: source,
	})
	if !ok {
		return relationQueryCacheValue{}, false
	}

	if typed, ok := val.(relationQueryCacheValue); ok {
		return typed, true
	}

	return relationQueryCacheValue{}, false
}

func (cache *relationQueryCache) Store(cat point.Category, source string, val relationQueryCacheValue) {
	if cache == nil {
		return
	}

	key := relationQueryCacheKey{
		cat:    cat,
		source: source,
	}

	if _, loaded := cache.data.LoadOrStore(key, val); loaded {
		return
	}

	if cache.size.Add(1) <= cache.limit {
		return
	}

	cache.data.Delete(key)
	cache.size.Add(^uint32(0))
}

func hasWildcard(source string) bool {
	for _, ch := range source {
		switch ch {
		case '*', '?':
			return true
		}
	}

	return false
}

func newWildcardRelation(pattern, name string) wildcardRelation {
	patternRunes := []rune(pattern)
	rule := wildcardRelation{
		pattern:      pattern,
		patternRunes: patternRunes,
		name:         name,
		patternLen:   len(patternRunes),
	}
	rule.prefix, rule.suffix = literalBoundary(pattern)

	for _, ch := range pattern {
		switch ch {
		case '*':
			rule.starCount++
		case '?':
			rule.minLen++
		default:
			rule.literalCount++
			rule.minLen++
		}
	}

	return rule
}

func (rule wildcardRelation) less(other wildcardRelation) bool {
	// More literal characters is more specific.
	if rule.literalCount != other.literalCount {
		return rule.literalCount > other.literalCount
	}

	// `?` is more restrictive than `*`, so fewer `*` wins.
	if rule.starCount != other.starCount {
		return rule.starCount < other.starCount
	}

	// Longer patterns constrain the source more tightly.
	if rule.patternLen != other.patternLen {
		return rule.patternLen > other.patternLen
	}

	return rule.pattern < other.pattern
}

func (index *wildcardIndex) add(rule wildcardRelation) {
	switch {
	case rule.prefix != "":
		index.prefixTrie = addWildcardTrieRule(index.prefixTrie, []rune(rule.prefix), rule)
	case rule.suffix != "":
		index.suffixTrie = addWildcardTrieRule(index.suffixTrie, reverseRunes([]rune(rule.suffix)), rule)
	default:
		if anchor := longestLiteralAnchor(rule.patternRunes); len(anchor) > 0 {
			if index.anchor == nil {
				index.anchor = newLiteralAnchorIndex()
			}
			index.anchor.add(rule, anchor)
			return
		}

		index.generic = append(index.generic, rule)
	}
}

func (index *wildcardIndex) hasRules() bool {
	return index != nil && (index.prefixTrie != nil || index.suffixTrie != nil || len(index.generic) > 0 || (index.anchor != nil && index.anchor.hasRules()))
}

func (index *wildcardIndex) finalize() {
	if index == nil {
		return
	}

	finalizeWildcardTrie(index.prefixTrie)
	finalizeWildcardTrie(index.suffixTrie)

	if index.anchor != nil {
		index.anchor.finalize()
	}
}

func (index *wildcardIndex) query(source string, sourceRunes []rune) (string, bool) {
	var (
		best    wildcardRelation
		matched bool
	)

	for _, rule := range index.generic {
		if rule.match(source, sourceRunes) {
			updateBestWildcard(rule, &best, &matched)
		}
	}

	if index.anchor != nil {
		index.anchor.matchAnchors(source, sourceRunes, &best, &matched)
	}

	queryWildcardTrie(index.prefixTrie, source, sourceRunes, false, &best, &matched)
	queryWildcardTrie(index.suffixTrie, source, sourceRunes, true, &best, &matched)

	if matched {
		return best.name, true
	}

	return "", false
}

func newLiteralAnchorIndex() *literalAnchorIndex {
	return &literalAnchorIndex{
		nodes: []literalAnchorNode{{}},
	}
}

func buildLiteralAnchorIndex(rel map[string]string) *literalAnchorIndex {
	idx := newLiteralAnchorIndex()
	idx.exact = make(map[string]string, len(rel))

	for pattern, name := range rel {
		if !hasWildcard(pattern) {
			idx.exact[pattern] = name
			continue
		}

		idx.add(newWildcardRelation(pattern, name), longestLiteralAnchor([]rune(pattern)))
	}

	idx.finalize()
	return idx
}

func (idx *literalAnchorIndex) add(rule wildcardRelation, anchor []rune) {
	if idx == nil {
		return
	}

	if len(anchor) == 0 {
		idx.generic = append(idx.generic, rule)
		return
	}

	anchorID := idx.addAnchor(anchor)
	idx.anchorRules[anchorID] = append(idx.anchorRules[anchorID], rule)
}

func (idx *literalAnchorIndex) hasRules() bool {
	return idx != nil && (len(idx.generic) > 0 || len(idx.anchorRules) > 0)
}

func (idx *literalAnchorIndex) finalize() {
	if idx == nil {
		return
	}

	idx.buildFailLinks()
	idx.scratchPool.New = func() interface{} {
		return &literalAnchorScratch{
			marks: make([]uint32, len(idx.anchorRules)),
		}
	}
}

func (idx *literalAnchorIndex) query(source string) (string, bool) {
	if idx == nil {
		return "", false
	}

	if name, ok := idx.exact[source]; ok {
		return name, true
	}

	sourceRunes := []rune(source)
	var (
		best    wildcardRelation
		matched bool
	)

	for _, rule := range idx.generic {
		if rule.match(source, sourceRunes) {
			updateBestWildcard(rule, &best, &matched)
		}
	}

	idx.matchAnchors(source, sourceRunes, &best, &matched)

	if matched {
		return best.name, true
	}

	return "", false
}

func (idx *literalAnchorIndex) addAnchor(anchor []rune) int {
	node := 0

	for _, ch := range anchor {
		if idx.nodes[node].children == nil {
			idx.nodes[node].children = map[rune]int{}
		}

		next, ok := idx.nodes[node].children[ch]
		if !ok {
			next = len(idx.nodes)
			idx.nodes = append(idx.nodes, literalAnchorNode{})
			idx.nodes[node].children[ch] = next
		}
		node = next
	}

	anchorID := len(idx.anchorRules)
	idx.anchorRules = append(idx.anchorRules, nil)
	idx.nodes[node].outputs = append(idx.nodes[node].outputs, anchorID)
	return anchorID
}

func (idx *literalAnchorIndex) buildFailLinks() {
	queue := make([]int, 0, len(idx.nodes))

	for _, child := range idx.nodes[0].children {
		queue = append(queue, child)
	}

	for head := 0; head < len(queue); head++ {
		cur := queue[head]

		for ch, next := range idx.nodes[cur].children {
			fail := idx.nodes[cur].fail
			for fail != 0 {
				if _, ok := idx.nodes[fail].children[ch]; ok {
					break
				}
				fail = idx.nodes[fail].fail
			}

			if target, ok := idx.nodes[fail].children[ch]; ok && target != next {
				idx.nodes[next].fail = target
			}

			idx.nodes[next].outputs = append(idx.nodes[next].outputs, idx.nodes[idx.nodes[next].fail].outputs...)
			queue = append(queue, next)
		}
	}
}

func (idx *literalAnchorIndex) matchAnchors(source string, sourceRunes []rune, best *wildcardRelation, matched *bool) {
	if idx == nil || len(idx.anchorRules) == 0 {
		return
	}

	scratch := idx.scratchPool.Get().(*literalAnchorScratch)
	if len(scratch.marks) < len(idx.anchorRules) {
		scratch.marks = make([]uint32, len(idx.anchorRules))
	}

	scratch.gen++
	if scratch.gen == 0 {
		for i := range scratch.marks {
			scratch.marks[i] = 0
		}
		scratch.gen = 1
	}

	node := 0
	for _, ch := range sourceRunes {
		for node != 0 {
			if _, ok := idx.nodes[node].children[ch]; ok {
				break
			}
			node = idx.nodes[node].fail
		}

		if next, ok := idx.nodes[node].children[ch]; ok {
			node = next
		}

		for _, anchorID := range idx.nodes[node].outputs {
			if scratch.marks[anchorID] == scratch.gen {
				continue
			}
			scratch.marks[anchorID] = scratch.gen

			for _, rule := range idx.anchorRules[anchorID] {
				if rule.match(source, sourceRunes) {
					updateBestWildcard(rule, best, matched)
				}
			}
		}
	}

	idx.scratchPool.Put(scratch)
}

func (rule wildcardRelation) match(source string, sourceRunes []rune) bool {
	if len(sourceRunes) < rule.minLen {
		return false
	}
	if rule.starCount == 0 && len(sourceRunes) != rule.patternLen {
		return false
	}
	if rule.prefix != "" && !strings.HasPrefix(source, rule.prefix) {
		return false
	}
	if rule.suffix != "" && !strings.HasSuffix(source, rule.suffix) {
		return false
	}

	return wildcardMatchRunes(rule.patternRunes, sourceRunes)
}

func queryWildcardMap(rel map[string]string, source string, sourceRunes []rune) (string, bool) {
	var (
		best    wildcardRelation
		matched bool
	)
	for pattern, name := range rel {
		if !hasWildcard(pattern) {
			continue
		}

		cur := newWildcardRelation(pattern, name)
		if !cur.match(source, sourceRunes) {
			continue
		}
		if !matched || cur.less(best) {
			best = cur
			matched = true
		}
	}

	if matched {
		return best.name, true
	}

	return "", false
}

func addWildcardTrieRule(root *wildcardTrieNode, key []rune, rule wildcardRelation) *wildcardTrieNode {
	if root == nil {
		root = &wildcardTrieNode{}
	}

	node := root
	for _, ch := range key {
		if node.children == nil {
			node.children = map[rune]*wildcardTrieNode{}
		}
		child := node.children[ch]
		if child == nil {
			child = &wildcardTrieNode{}
			node.children[ch] = child
		}
		node = child
	}

	node.rules = append(node.rules, rule)
	return root
}

func finalizeWildcardTrie(root *wildcardTrieNode) {
	if root == nil {
		return
	}

	if len(root.rules) > 0 {
		root.exactLen = map[int][]wildcardRelation{}
		root.starByLen = map[int][]wildcardRelation{}
		root.starLens = root.starLens[:0]

		for _, rule := range root.rules {
			if rule.starCount == 0 {
				root.exactLen[rule.patternLen] = append(root.exactLen[rule.patternLen], rule)
				continue
			}

			if _, ok := root.starByLen[rule.minLen]; !ok {
				root.starLens = append(root.starLens, rule.minLen)
			}
			root.starByLen[rule.minLen] = append(root.starByLen[rule.minLen], rule)
		}

		sort.Ints(root.starLens)
		root.indexed = true
	}

	for _, child := range root.children {
		finalizeWildcardTrie(child)
	}
}

func queryWildcardTrie(root *wildcardTrieNode, source string, sourceRunes []rune, reverse bool, best *wildcardRelation, matched *bool) {
	if root == nil {
		return
	}

	node := root
	if reverse {
		for i := len(sourceRunes) - 1; i >= 0; i-- {
			node = node.children[sourceRunes[i]]
			if node == nil {
				return
			}
			queryWildcardTrieNode(node, source, sourceRunes, best, matched)
		}
		return
	}

	for _, ch := range sourceRunes {
		node = node.children[ch]
		if node == nil {
			return
		}
		queryWildcardTrieNode(node, source, sourceRunes, best, matched)
	}
}

func queryWildcardTrieNode(node *wildcardTrieNode, source string, sourceRunes []rune, best *wildcardRelation, matched *bool) {
	if node == nil || len(node.rules) == 0 {
		return
	}

	if !node.indexed {
		for _, rule := range node.rules {
			if rule.match(source, sourceRunes) {
				updateBestWildcard(rule, best, matched)
			}
		}
		return
	}

	n := len(sourceRunes)
	for _, rule := range node.exactLen[n] {
		if rule.match(source, sourceRunes) {
			updateBestWildcard(rule, best, matched)
		}
	}

	for _, minLen := range node.starLens {
		if minLen > n {
			break
		}
		for _, rule := range node.starByLen[minLen] {
			if rule.match(source, sourceRunes) {
				updateBestWildcard(rule, best, matched)
			}
		}
	}
}

func updateBestWildcard(rule wildcardRelation, best *wildcardRelation, matched *bool) {
	if !*matched || rule.less(*best) {
		*best = rule
		*matched = true
	}
}

func longestLiteralAnchor(pattern []rune) []rune {
	var (
		best      []rune
		curStart  = -1
		curLen    int
		bestStart int
		bestLen   int
	)

	flush := func() {
		if curLen > bestLen {
			bestStart = curStart
			bestLen = curLen
		}
		curStart = -1
		curLen = 0
	}

	for i, ch := range pattern {
		switch ch {
		case '*', '?':
			if curLen > 0 {
				flush()
			}
		default:
			if curStart < 0 {
				curStart = i
			}
			curLen++
		}
	}

	if curLen > 0 {
		flush()
	}

	if bestLen == 0 {
		return nil
	}

	best = make([]rune, bestLen)
	copy(best, pattern[bestStart:bestStart+bestLen])
	return best
}

func literalBoundary(pattern string) (prefix, suffix string) {
	first := strings.IndexAny(pattern, "*?")
	if first < 0 {
		return pattern, pattern
	}

	if first > 0 {
		prefix = pattern[:first]
	}

	last := strings.LastIndexAny(pattern, "*?")
	if last >= 0 && last+1 < len(pattern) {
		suffix = pattern[last+1:]
	}

	return prefix, suffix
}

func reverseRunes(r []rune) []rune {
	if len(r) == 0 {
		return nil
	}

	rev := make([]rune, len(r))
	for i := range r {
		rev[len(r)-1-i] = r[i]
	}
	return rev
}

func wildcardMatchRunes(pat, src []rune) bool {
	pIdx, sIdx := 0, 0
	starIdx, matchIdx := -1, 0

	for sIdx < len(src) {
		switch {
		case pIdx < len(pat) && (pat[pIdx] == '?' || pat[pIdx] == src[sIdx]):
			pIdx++
			sIdx++
		case pIdx < len(pat) && pat[pIdx] == '*':
			starIdx = pIdx
			matchIdx = sIdx
			pIdx++
		case starIdx >= 0:
			pIdx = starIdx + 1
			matchIdx++
			sIdx = matchIdx
		default:
			return false
		}
	}

	for pIdx < len(pat) && pat[pIdx] == '*' {
		pIdx++
	}

	return pIdx == len(pat)
}
