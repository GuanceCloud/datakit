package scan

import (
	"sort"
	"strings"
)

type Scanner struct {
	atomCount int
	useTrie   bool
	buckets   [256][]entry
	rootNext  [256]int
	nodes     []trieNode
}

type entry struct {
	id  int
	key string
}

type trieNode struct {
	outputs []int
	edges   []trieEdge
	fail    int
}

type trieEdge struct {
	b    byte
	next int
}

type Result struct {
	hits []bool
}

func New(atoms []string) *Scanner {
	if len(atoms) == 0 {
		return nil
	}

	s := &Scanner{
		atomCount: len(atoms),
		useTrie:   shouldUseTrieScanner(atoms),
	}
	if s.useTrie {
		s.initTrie(atoms)
	} else {
		s.initBuckets(atoms)
	}
	return s
}

func (s *Scanner) initBuckets(atoms []string) {
	for i, atom := range atoms {
		if atom == "" {
			continue
		}
		s.buckets[atom[0]] = append(s.buckets[atom[0]], entry{
			id:  i,
			key: atom,
		})
	}
}

func (s *Scanner) initTrie(atoms []string) {
	s.nodes = make([]trieNode, 1, len(atoms)+1)
	for i := range s.rootNext {
		s.rootNext[i] = -1
	}

	for id, atom := range atoms {
		if atom == "" {
			continue
		}

		node := s.ensureRootNode(atom[0])
		for i := 1; i < len(atom); i++ {
			node = s.ensureChildNode(node, atom[i])
		}
		s.nodes[node].outputs = append(s.nodes[node].outputs, id)
	}
	s.buildTrieFailureLinks()
}

func (s *Scanner) Scan(content string) Result {
	result := Result{
		hits: make([]bool, s.atomCount),
	}
	s.ScanInto(content, result.hits)
	return result
}

func (s *Scanner) ScanInto(content string, hits []bool) {
	if s == nil || len(content) == 0 || len(hits) == 0 {
		return
	}
	if s.useTrie {
		s.scanTrieInto(content, hits)
		return
	}
	s.scanBucketsInto(content, hits)
}

func (s *Scanner) ScanBits(content string) uint64 {
	if s == nil || len(content) == 0 || s.atomCount == 0 || s.atomCount > 64 {
		return 0
	}
	if s.useTrie {
		return s.scanTrieBits(content)
	}
	return s.scanBucketBits(content)
}

func (s *Scanner) scanBucketsInto(content string, hits []bool) {
	for i := 0; i < len(content); i++ {
		bucket := s.buckets[content[i]]
		if len(bucket) == 0 {
			continue
		}

		remaining := content[i:]
		for _, entry := range bucket {
			if hits[entry.id] || len(entry.key) > len(remaining) {
				continue
			}
			if strings.HasPrefix(remaining, entry.key) {
				hits[entry.id] = true
			}
		}
	}
}

func (s *Scanner) scanBucketBits(content string) uint64 {
	var bits uint64
	for i := 0; i < len(content); i++ {
		bucket := s.buckets[content[i]]
		if len(bucket) == 0 {
			continue
		}

		remaining := content[i:]
		for _, entry := range bucket {
			mask := uint64(1) << entry.id
			if bits&mask != 0 || len(entry.key) > len(remaining) {
				continue
			}
			if strings.HasPrefix(remaining, entry.key) {
				bits |= mask
			}
		}
	}
	return bits
}

func (s *Scanner) scanTrieInto(content string, hits []bool) {
	node := 0
	for i := 0; i < len(content); i++ {
		node = s.advanceTrieState(node, content[i])
		if node <= 0 {
			continue
		}
		s.markOutputs(node, hits)
	}
}

func (s *Scanner) scanTrieBits(content string) uint64 {
	var bits uint64
	node := 0
	for i := 0; i < len(content); i++ {
		node = s.advanceTrieState(node, content[i])
		if node <= 0 {
			continue
		}
		bits = s.markOutputBits(node, bits)
	}
	return bits
}

func shouldUseTrieScanner(atoms []string) bool {
	if len(atoms) < 4 {
		return false
	}

	sorted := append([]string(nil), atoms...)
	sort.Strings(sorted)

	sharedScore := 0
	sharedPairs := 0
	for i := 1; i < len(sorted); i++ {
		prefix := sharedPrefixLen(sorted[i-1], sorted[i])
		if prefix < 2 {
			continue
		}
		sharedPairs++
		sharedScore += prefix
	}
	if sharedPairs == 0 {
		return false
	}
	return sharedScore >= len(atoms)
}

func sharedPrefixLen(left, right string) int {
	n := len(left)
	if len(right) < n {
		n = len(right)
	}
	i := 0
	for i < n && left[i] == right[i] {
		i++
	}
	return i
}

func (s *Scanner) ensureRootNode(b byte) int {
	node := s.rootNext[b]
	if node >= 0 {
		return node
	}
	node = len(s.nodes)
	s.nodes = append(s.nodes, trieNode{})
	s.rootNext[b] = node
	return node
}

func (s *Scanner) ensureChildNode(parent int, b byte) int {
	edges := s.nodes[parent].edges
	for i := range edges {
		if edges[i].b == b {
			return edges[i].next
		}
	}

	child := len(s.nodes)
	s.nodes = append(s.nodes, trieNode{})
	s.nodes[parent].edges = append(s.nodes[parent].edges, trieEdge{
		b:    b,
		next: child,
	})
	return child
}

func (s *Scanner) nextNode(node int, b byte) int {
	if node <= 0 || node >= len(s.nodes) {
		return -1
	}
	for _, edge := range s.nodes[node].edges {
		if edge.b == b {
			return edge.next
		}
	}
	return -1
}

func (s *Scanner) advanceTrieState(node int, b byte) int {
	if node == 0 {
		return s.rootNext[b]
	}

	for node > 0 {
		if next := s.nextNode(node, b); next >= 0 {
			return next
		}
		node = s.nodes[node].fail
	}
	return s.rootNext[b]
}

func (s *Scanner) buildTrieFailureLinks() {
	queue := make([]int, 0, len(s.nodes))
	for b := 0; b < len(s.rootNext); b++ {
		node := s.rootNext[byte(b)]
		if node < 0 {
			continue
		}
		s.nodes[node].fail = 0
		queue = append(queue, node)
	}

	for head := 0; head < len(queue); head++ {
		node := queue[head]
		for _, edge := range s.nodes[node].edges {
			fail := s.nodes[node].fail
			for fail > 0 {
				if next := s.nextNode(fail, edge.b); next >= 0 {
					fail = next
					break
				}
				fail = s.nodes[fail].fail
			}
			if fail == 0 {
				if next := s.rootNext[edge.b]; next >= 0 && next != edge.next {
					fail = next
				}
			}
			s.nodes[edge.next].fail = fail
			if len(s.nodes[fail].outputs) > 0 {
				s.nodes[edge.next].outputs = append(s.nodes[edge.next].outputs, s.nodes[fail].outputs...)
			}
			queue = append(queue, edge.next)
		}
	}
}

func (s *Scanner) markOutputs(node int, hits []bool) {
	for _, id := range s.nodes[node].outputs {
		if id >= 0 && id < len(hits) {
			hits[id] = true
		}
	}
}

func (s *Scanner) markOutputBits(node int, bits uint64) uint64 {
	for _, id := range s.nodes[node].outputs {
		if id >= 0 && id < 64 {
			bits |= uint64(1) << id
		}
	}
	return bits
}

func (r Result) Has(id int) bool {
	return id >= 0 && id < len(r.hits) && r.hits[id]
}
