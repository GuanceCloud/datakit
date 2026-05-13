package grok

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type grokIRKind uint8

const (
	grokIREmpty grokIRKind = iota
	grokIRSequence
	grokIRAlternate
	grokIRLiteral
	grokIRRef
	grokIRRawRegex
	grokIRGroup
	grokIRRepeat
)

type grokIRGroupKind uint8

const (
	grokIRGroupCapture grokIRGroupKind = iota
	grokIRGroupNonCapture
)

type grokIRPattern struct {
	source string
	root   *grokIRNode
}

type grokIRNode struct {
	kind      grokIRKind
	children  []*grokIRNode
	literal   string
	raw       string
	ref       patternRef
	groupKind grokIRGroupKind
	min       int
	max       int
	greedy    bool
}

type matchIRKind uint8

const (
	matchIREmpty matchIRKind = iota
	matchIRSequence
	matchIRAlternate
	matchIRLiteral
	matchIRRef
	matchIRRawRegex
	matchIRRepeat
)

type matchIRPattern struct {
	source string
	root   *matchIRNode
}

type matchIRNode struct {
	kind     matchIRKind
	children []*matchIRNode
	literal  string
	raw      string
	ref      patternRef
	min      int
	max      int
	greedy   bool
}

type patternPrefilter struct {
	prefix    string
	suffix    string
	literals  map[string]struct{}
	minLength int
	exact     string
}

var (
	compiledGrokIRCache  sync.Map
	compiledMatchIRCache sync.Map
)

func compileGrokIR(pattern string) (*grokIRPattern, error) {
	if cached, ok := compiledGrokIRCache.Load(pattern); ok {
		return cached.(*grokIRPattern), nil
	}

	p := grokIRParser{pattern: pattern}
	root, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}
	if p.pos != len(pattern) {
		return nil, fmt.Errorf("unexpected trailing input at %d", p.pos)
	}

	compiled := &grokIRPattern{
		source: pattern,
		root:   root,
	}
	actual, _ := compiledGrokIRCache.LoadOrStore(pattern, compiled)
	return actual.(*grokIRPattern), nil
}

func compileMatchIR(pattern string) (*matchIRPattern, error) {
	if cached, ok := compiledMatchIRCache.Load(pattern); ok {
		return cached.(*matchIRPattern), nil
	}

	grokPattern, err := compileGrokIR(pattern)
	if err != nil {
		return nil, err
	}

	compiled := &matchIRPattern{
		source: pattern,
		root:   buildMatchIR(grokPattern.root),
	}
	actual, _ := compiledMatchIRCache.LoadOrStore(pattern, compiled)
	return actual.(*matchIRPattern), nil
}

func compilePatternPrefilterIR(pattern string, storage PatternStorageIface) (string, string, []string, int) {
	compiled, err := compileMatchIR(pattern)
	if err != nil {
		return "", "", nil, 0
	}
	info := compilePatternPrefilterFromMatchIR(compiled.root, storage)
	return info.prefix, info.suffix, finalizePrefilterLiterals(info), info.minLength
}

func compilePatternPrefilterFromMatchIR(root *matchIRNode, storage PatternStorageIface) patternPrefilter {
	strippedRoot, _, anchoredEnd := stripMatchIRBoundaryAnchors(root)

	cache := make(map[string]patternPrefilter)
	visiting := make(map[string]bool)
	info := deriveMatchIRPrefilter(strippedRoot, storage, cache, visiting)
	if !anchoredEnd {
		info.suffix = ""
	}
	return info
}

func dumpGrokIR(pattern string) (string, error) {
	compiled, err := compileGrokIR(pattern)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("GrokIR\n")
	dumpGrokIRNode(&b, compiled.root, 0)
	return b.String(), nil
}

func dumpMatchIR(pattern string) (string, error) {
	compiled, err := compileMatchIR(pattern)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("MatchIR\n")
	dumpMatchIRNode(&b, compiled.root, 0)
	return b.String(), nil
}

func deriveMatchIRPrefilter(node *matchIRNode, storage PatternStorageIface, cache map[string]patternPrefilter, visiting map[string]bool) patternPrefilter {
	if node == nil {
		return patternPrefilter{}
	}

	switch node.kind {
	case matchIREmpty:
		return patternPrefilter{}
	case matchIRLiteral:
		return patternPrefilter{
			prefix:    node.literal,
			suffix:    node.literal,
			literals:  literalSet(node.literal),
			minLength: len(node.literal),
			exact:     node.literal,
		}
	case matchIRRef:
		return deriveRefPrefilter(node.ref.syntax, storage, cache, visiting)
	case matchIRRawRegex:
		return deriveRawRegexPrefilter(node.raw)
	case matchIRSequence:
		return deriveSequencePrefilter(node.children, storage, cache, visiting)
	case matchIRAlternate:
		return deriveAlternatePrefilter(node.children, storage, cache, visiting)
	case matchIRRepeat:
		return deriveRepeatPrefilter(node, storage, cache, visiting)
	default:
		return patternPrefilter{}
	}
}

func deriveRefPrefilter(name string, storage PatternStorageIface, cache map[string]patternPrefilter, visiting map[string]bool) patternPrefilter {
	if name == "" || storage == nil {
		return patternPrefilter{}
	}
	if info, ok := cache[name]; ok {
		return info
	}
	if visiting[name] {
		return patternPrefilter{}
	}

	gp, ok := storage.GetPattern(name)
	if !ok || gp == nil {
		return patternPrefilter{}
	}

	visiting[name] = true
	defer delete(visiting, name)

	child, err := compileMatchIR(gp.pattern)
	if err != nil {
		return patternPrefilter{}
	}
	info := deriveMatchIRPrefilter(child.root, storage, cache, visiting)
	cache[name] = info
	return info
}

func deriveRawRegexPrefilter(raw string) patternPrefilter {
	switch raw {
	case "", "^", "$", `\b`, `\B`, `\A`, `\z`, `(?s)`:
		return patternPrefilter{literals: map[string]struct{}{}}
	case ".", `\s`, `\S`, `\d`, `\D`, `\w`, `\W`:
		return patternPrefilter{literals: map[string]struct{}{}, minLength: 1}
	default:
		if strings.HasPrefix(raw, "[") && strings.HasSuffix(raw, "]") && len(raw) >= 2 {
			return patternPrefilter{literals: map[string]struct{}{}, minLength: 1}
		}
		return patternPrefilter{literals: map[string]struct{}{}}
	}
}

func deriveSequencePrefilter(children []*matchIRNode, storage PatternStorageIface, cache map[string]patternPrefilter, visiting map[string]bool) patternPrefilter {
	info := patternPrefilter{literals: map[string]struct{}{}}
	var prefix strings.Builder
	var suffix strings.Builder
	prefixOpen := true
	suffixOpen := true
	childInfos := make([]patternPrefilter, len(children))

	for i, child := range children {
		childInfo := deriveMatchIRPrefilter(child, storage, cache, visiting)
		childInfos[i] = childInfo
		info.minLength += childInfo.minLength
		unionLiteralSets(info.literals, childInfo.literals)

		if !prefixOpen {
			continue
		}
		if childInfo.exact != "" {
			prefix.WriteString(childInfo.exact)
			continue
		}
		prefix.WriteString(childInfo.prefix)
		prefixOpen = false
	}
	for i := len(childInfos) - 1; i >= 0; i-- {
		childInfo := childInfos[i]
		if !suffixOpen {
			continue
		}
		if childInfo.exact != "" {
			suffix.WriteString(childInfo.exact)
			continue
		}
		suffix.WriteString(childInfo.suffix)
		suffixOpen = false
	}

	info.prefix = prefix.String()
	info.suffix = reverseString(suffix.String())
	if info.minLength == len(info.prefix) && allChildPrefiltersExact(childInfos) {
		info.exact = info.prefix
	}
	if info.exact != "" {
		info.suffix = info.exact
		info.literals[info.exact] = struct{}{}
	}
	return info
}

func allChildPrefiltersExact(children []patternPrefilter) bool {
	for _, child := range children {
		if child.exact == "" {
			return false
		}
	}
	return true
}

func deriveAlternatePrefilter(children []*matchIRNode, storage PatternStorageIface, cache map[string]patternPrefilter, visiting map[string]bool) patternPrefilter {
	if len(children) == 0 {
		return patternPrefilter{}
	}

	first := deriveMatchIRPrefilter(children[0], storage, cache, visiting)
	prefix := first.prefix
	suffix := first.suffix
	minLength := first.minLength
	exact := first.exact
	literals := copyLiteralSet(first.literals)

	for _, child := range children[1:] {
		childInfo := deriveMatchIRPrefilter(child, storage, cache, visiting)
		prefix = commonPrefix(prefix, childInfo.prefix)
		suffix = commonSuffix(suffix, childInfo.suffix)
		if childInfo.minLength < minLength {
			minLength = childInfo.minLength
		}
		literals = intersectLiteralSets(literals, childInfo.literals)
		if exact == "" || exact != childInfo.exact {
			exact = ""
		}
	}

	return patternPrefilter{
		prefix:    prefix,
		suffix:    suffix,
		literals:  literals,
		minLength: minLength,
		exact:     exact,
	}
}

func deriveRepeatPrefilter(node *matchIRNode, storage PatternStorageIface, cache map[string]patternPrefilter, visiting map[string]bool) patternPrefilter {
	if len(node.children) == 0 {
		return patternPrefilter{}
	}
	child := deriveMatchIRPrefilter(node.children[0], storage, cache, visiting)
	if node.min == 0 {
		return patternPrefilter{}
	}

	info := patternPrefilter{
		literals:  copyLiteralSet(child.literals),
		minLength: child.minLength * node.min,
	}

	if child.exact != "" {
		info.prefix = strings.Repeat(child.exact, node.min)
		info.suffix = info.prefix
		info.literals = literalSet(child.exact)
		if node.min > 1 {
			info.literals[info.prefix] = struct{}{}
		}
		if node.max == node.min {
			info.exact = info.prefix
		}
		return info
	}

	info.prefix = child.prefix
	info.suffix = child.suffix
	return info
}

func finalizePrefilterLiterals(info patternPrefilter) []string {
	if len(info.literals) == 0 {
		return nil
	}

	out := make([]string, 0, len(info.literals))
	for lit := range info.literals {
		if len(lit) < 2 || lit == info.prefix {
			continue
		}
		out = append(out, lit)
	}
	if len(out) <= 1 {
		return out
	}

	sort.Slice(out, func(i, j int) bool {
		if len(out[i]) == len(out[j]) {
			return out[i] < out[j]
		}
		return len(out[i]) > len(out[j])
	})
	if len(out) > 2 {
		out = out[:2]
	}
	return out
}

type grokIRParser struct {
	pattern string
	pos     int
}

func (p *grokIRParser) parseExpression(stop byte) (*grokIRNode, error) {
	alternatives := make([]*grokIRNode, 0, 2)

	for {
		seq, err := p.parseSequence(stop)
		if err != nil {
			return nil, err
		}
		alternatives = append(alternatives, seq)

		if p.pos >= len(p.pattern) || (stop != 0 && p.pattern[p.pos] == stop) {
			break
		}
		if p.pattern[p.pos] != '|' {
			return nil, fmt.Errorf("unexpected token %q at %d", p.pattern[p.pos], p.pos)
		}
		p.pos++
	}

	return normalizeGrokAlternates(alternatives), nil
}

func (p *grokIRParser) parseSequence(stop byte) (*grokIRNode, error) {
	children := make([]*grokIRNode, 0, 4)

	for p.pos < len(p.pattern) {
		if stop != 0 && p.pattern[p.pos] == stop {
			break
		}
		if p.pattern[p.pos] == '|' {
			break
		}

		node, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		children = append(children, node)
	}

	return normalizeGrokSequence(children), nil
}

func (p *grokIRParser) parseTerm() (*grokIRNode, error) {
	atom, err := p.parseAtom()
	if err != nil {
		return nil, err
	}
	return p.parseQuantifier(atom)
}

func (p *grokIRParser) parseAtom() (*grokIRNode, error) {
	if p.pos >= len(p.pattern) {
		return emptyGrokNode(), nil
	}

	switch {
	case strings.HasPrefix(p.pattern[p.pos:], "%{"):
		return p.parseRef()
	case p.pattern[p.pos] == '(':
		return p.parseGroup()
	case p.pattern[p.pos] == '[':
		return p.parseCharClass()
	case p.pattern[p.pos] == '\\':
		return p.parseEscape()
	case p.pattern[p.pos] == '.', p.pattern[p.pos] == '^', p.pattern[p.pos] == '$':
		raw := p.pattern[p.pos : p.pos+1]
		p.pos++
		return &grokIRNode{kind: grokIRRawRegex, raw: raw}, nil
	case isLiteralByte(p.pattern[p.pos]):
		return p.parseLiteral()
	default:
		return nil, fmt.Errorf("unexpected token %q at %d", p.pattern[p.pos], p.pos)
	}
}

func (p *grokIRParser) parseRef() (*grokIRNode, error) {
	end := strings.IndexByte(p.pattern[p.pos+2:], '}')
	if end < 0 {
		return nil, fmt.Errorf("unterminated grok reference at %d", p.pos)
	}
	end += p.pos + 2

	ref, ok, err := parsePatternRef(p.pattern[p.pos+2 : end])
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("invalid grok reference at %d", p.pos)
	}

	p.pos = end + 1
	return &grokIRNode{kind: grokIRRef, ref: ref}, nil
}

func (p *grokIRParser) parseGroup() (*grokIRNode, error) {
	start := p.pos

	switch {
	case strings.HasPrefix(p.pattern[p.pos:], "(?:"):
		p.pos += 3
		child, err := p.parseExpression(')')
		if err != nil {
			return nil, err
		}
		if p.pos >= len(p.pattern) || p.pattern[p.pos] != ')' {
			return nil, fmt.Errorf("unterminated non-capturing group at %d", start)
		}
		p.pos++
		return &grokIRNode{
			kind:      grokIRGroup,
			groupKind: grokIRGroupNonCapture,
			children:  []*grokIRNode{child},
		}, nil
	case strings.HasPrefix(p.pattern[p.pos:], "(?"):
		end, err := scanBalancedPatternGroup(p.pattern, p.pos)
		if err != nil {
			return nil, err
		}
		raw := p.pattern[p.pos : end+1]
		p.pos = end + 1
		return &grokIRNode{kind: grokIRRawRegex, raw: raw}, nil
	default:
		p.pos++
		child, err := p.parseExpression(')')
		if err != nil {
			return nil, err
		}
		if p.pos >= len(p.pattern) || p.pattern[p.pos] != ')' {
			return nil, fmt.Errorf("unterminated capturing group at %d", start)
		}
		p.pos++
		return &grokIRNode{
			kind:      grokIRGroup,
			groupKind: grokIRGroupCapture,
			children:  []*grokIRNode{child},
		}, nil
	}
}

func (p *grokIRParser) parseCharClass() (*grokIRNode, error) {
	start := p.pos
	p.pos++
	for p.pos < len(p.pattern) {
		switch p.pattern[p.pos] {
		case '\\':
			if p.pos+1 >= len(p.pattern) {
				return nil, fmt.Errorf("unterminated character class at %d", start)
			}
			p.pos += 2
		case ']':
			p.pos++
			return &grokIRNode{kind: grokIRRawRegex, raw: p.pattern[start:p.pos]}, nil
		default:
			p.pos++
		}
	}
	return nil, fmt.Errorf("unterminated character class at %d", start)
}

func (p *grokIRParser) parseEscape() (*grokIRNode, error) {
	if p.pos+1 >= len(p.pattern) {
		return nil, fmt.Errorf("dangling escape at %d", p.pos)
	}

	next := p.pattern[p.pos+1]
	p.pos += 2

	switch next {
	case 't':
		return &grokIRNode{kind: grokIRLiteral, literal: "\t"}, nil
	case 'n':
		return &grokIRNode{kind: grokIRLiteral, literal: "\n"}, nil
	case 'r':
		return &grokIRNode{kind: grokIRLiteral, literal: "\r"}, nil
	case 's', 'S', 'd', 'D', 'w', 'W', 'b', 'B', 'A', 'z', 'Z':
		return &grokIRNode{kind: grokIRRawRegex, raw: `\` + string(next)}, nil
	default:
		if isEscapedLiteralByte(next) {
			return &grokIRNode{kind: grokIRLiteral, literal: string(next)}, nil
		}
		return &grokIRNode{kind: grokIRRawRegex, raw: `\` + string(next)}, nil
	}
}

func (p *grokIRParser) parseLiteral() (*grokIRNode, error) {
	start := p.pos
	for p.pos < len(p.pattern) {
		if strings.HasPrefix(p.pattern[p.pos:], "%{") {
			break
		}
		if !isLiteralByte(p.pattern[p.pos]) {
			break
		}
		p.pos++
	}
	return &grokIRNode{kind: grokIRLiteral, literal: p.pattern[start:p.pos]}, nil
}

func (p *grokIRParser) parseQuantifier(atom *grokIRNode) (*grokIRNode, error) {
	if p.pos >= len(p.pattern) {
		return atom, nil
	}

	min := 1
	max := 1
	hasQuantifier := true

	switch p.pattern[p.pos] {
	case '?':
		min, max = 0, 1
		p.pos++
	case '*':
		min, max = 0, -1
		p.pos++
	case '+':
		min, max = 1, -1
		p.pos++
	case '{':
		parsedMin, parsedMax, nextPos, err := parseRepeatBounds(p.pattern, p.pos)
		if err != nil {
			return nil, err
		}
		min, max = parsedMin, parsedMax
		p.pos = nextPos
	default:
		hasQuantifier = false
	}

	if !hasQuantifier {
		return atom, nil
	}

	greedy := true
	if p.pos < len(p.pattern) && p.pattern[p.pos] == '?' {
		greedy = false
		p.pos++
	}

	return &grokIRNode{
		kind:     grokIRRepeat,
		children: []*grokIRNode{atom},
		min:      min,
		max:      max,
		greedy:   greedy,
	}, nil
}

func parseRepeatBounds(pattern string, pos int) (int, int, int, error) {
	start := pos
	pos++

	minStart := pos
	for pos < len(pattern) && isASCIIDigit(pattern[pos]) {
		pos++
	}
	if minStart == pos {
		return 0, 0, 0, fmt.Errorf("invalid repeat at %d", start)
	}

	min, err := strconv.Atoi(pattern[minStart:pos])
	if err != nil {
		return 0, 0, 0, err
	}

	max := min
	if pos < len(pattern) && pattern[pos] == ',' {
		pos++
		if pos < len(pattern) && pattern[pos] == '}' {
			max = -1
		} else {
			maxStart := pos
			for pos < len(pattern) && isASCIIDigit(pattern[pos]) {
				pos++
			}
			if maxStart == pos {
				return 0, 0, 0, fmt.Errorf("invalid repeat at %d", start)
			}
			max, err = strconv.Atoi(pattern[maxStart:pos])
			if err != nil {
				return 0, 0, 0, err
			}
		}
	}

	if pos >= len(pattern) || pattern[pos] != '}' {
		return 0, 0, 0, fmt.Errorf("unterminated repeat at %d", start)
	}
	pos++

	return min, max, pos, nil
}

func scanBalancedPatternGroup(pattern string, start int) (int, error) {
	depth := 0
	inClass := false

	for i := start; i < len(pattern); i++ {
		switch pattern[i] {
		case '\\':
			if i+1 < len(pattern) {
				i++
			}
		case '[':
			if !inClass {
				inClass = true
			}
		case ']':
			if inClass {
				inClass = false
			}
		case '(':
			if !inClass {
				depth++
			}
		case ')':
			if !inClass {
				depth--
				if depth == 0 {
					return i, nil
				}
			}
		}
	}

	return 0, fmt.Errorf("unterminated special group at %d", start)
}

func normalizeGrokSequence(children []*grokIRNode) *grokIRNode {
	switch len(children) {
	case 0:
		return emptyGrokNode()
	case 1:
		return children[0]
	default:
		return &grokIRNode{kind: grokIRSequence, children: children}
	}
}

func normalizeGrokAlternates(alternatives []*grokIRNode) *grokIRNode {
	switch len(alternatives) {
	case 0:
		return emptyGrokNode()
	case 1:
		return alternatives[0]
	default:
		return &grokIRNode{kind: grokIRAlternate, children: alternatives}
	}
}

func emptyGrokNode() *grokIRNode {
	return &grokIRNode{kind: grokIREmpty}
}

func buildMatchIR(node *grokIRNode) *matchIRNode {
	if node == nil {
		return &matchIRNode{kind: matchIREmpty}
	}

	switch node.kind {
	case grokIREmpty:
		return &matchIRNode{kind: matchIREmpty}
	case grokIRSequence:
		children := make([]*matchIRNode, 0, len(node.children))
		for _, child := range node.children {
			children = append(children, buildMatchIR(child))
		}
		return normalizeMatchSequence(children)
	case grokIRAlternate:
		children := make([]*matchIRNode, 0, len(node.children))
		for _, child := range node.children {
			children = append(children, buildMatchIR(child))
		}
		return normalizeMatchAlternates(children)
	case grokIRLiteral:
		return &matchIRNode{kind: matchIRLiteral, literal: node.literal}
	case grokIRRef:
		return &matchIRNode{kind: matchIRRef, ref: node.ref}
	case grokIRRawRegex:
		return &matchIRNode{kind: matchIRRawRegex, raw: node.raw}
	case grokIRGroup:
		if len(node.children) == 0 {
			return &matchIRNode{kind: matchIREmpty}
		}
		return buildMatchIR(node.children[0])
	case grokIRRepeat:
		child := &matchIRNode{kind: matchIREmpty}
		if len(node.children) > 0 {
			child = buildMatchIR(node.children[0])
		}
		return &matchIRNode{
			kind:     matchIRRepeat,
			children: []*matchIRNode{child},
			min:      node.min,
			max:      node.max,
			greedy:   node.greedy,
		}
	default:
		return &matchIRNode{kind: matchIREmpty}
	}
}

func normalizeMatchSequence(children []*matchIRNode) *matchIRNode {
	switch len(children) {
	case 0:
		return &matchIRNode{kind: matchIREmpty}
	case 1:
		return children[0]
	default:
		return &matchIRNode{kind: matchIRSequence, children: children}
	}
}

func normalizeMatchAlternates(children []*matchIRNode) *matchIRNode {
	switch len(children) {
	case 0:
		return &matchIRNode{kind: matchIREmpty}
	case 1:
		return children[0]
	default:
		return &matchIRNode{kind: matchIRAlternate, children: children}
	}
}

func dumpGrokIRNode(b *strings.Builder, node *grokIRNode, depth int) {
	writeIRIndent(b, depth)
	switch node.kind {
	case grokIREmpty:
		b.WriteString("Empty\n")
	case grokIRSequence:
		b.WriteString("Sequence\n")
		for _, child := range node.children {
			dumpGrokIRNode(b, child, depth+1)
		}
	case grokIRAlternate:
		b.WriteString("Alternate\n")
		for _, child := range node.children {
			dumpGrokIRNode(b, child, depth+1)
		}
	case grokIRLiteral:
		fmt.Fprintf(b, "Literal %q\n", node.literal)
	case grokIRRef:
		fmt.Fprintf(b, "Ref syntax=%q alias=%q type=%q\n", node.ref.syntax, node.ref.alias, node.ref.varType)
	case grokIRRawRegex:
		fmt.Fprintf(b, "RawRegex %q\n", node.raw)
	case grokIRGroup:
		fmt.Fprintf(b, "Group kind=%q\n", grokIRGroupKindString(node.groupKind))
		for _, child := range node.children {
			dumpGrokIRNode(b, child, depth+1)
		}
	case grokIRRepeat:
		fmt.Fprintf(b, "Repeat min=%d max=%s greedy=%t\n", node.min, formatRepeatMax(node.max), node.greedy)
		for _, child := range node.children {
			dumpGrokIRNode(b, child, depth+1)
		}
	}
}

func dumpMatchIRNode(b *strings.Builder, node *matchIRNode, depth int) {
	writeIRIndent(b, depth)
	switch node.kind {
	case matchIREmpty:
		b.WriteString("Empty\n")
	case matchIRSequence:
		b.WriteString("Sequence\n")
		for _, child := range node.children {
			dumpMatchIRNode(b, child, depth+1)
		}
	case matchIRAlternate:
		b.WriteString("Alternate\n")
		for _, child := range node.children {
			dumpMatchIRNode(b, child, depth+1)
		}
	case matchIRLiteral:
		fmt.Fprintf(b, "Literal %q\n", node.literal)
	case matchIRRef:
		fmt.Fprintf(b, "Ref syntax=%q alias=%q type=%q\n", node.ref.syntax, node.ref.alias, node.ref.varType)
	case matchIRRawRegex:
		fmt.Fprintf(b, "RawRegex %q\n", node.raw)
	case matchIRRepeat:
		fmt.Fprintf(b, "Repeat min=%d max=%s greedy=%t\n", node.min, formatRepeatMax(node.max), node.greedy)
		for _, child := range node.children {
			dumpMatchIRNode(b, child, depth+1)
		}
	}
}

func writeIRIndent(b *strings.Builder, depth int) {
	for i := 0; i < depth; i++ {
		b.WriteString("  ")
	}
}

func grokIRGroupKindString(kind grokIRGroupKind) string {
	switch kind {
	case grokIRGroupCapture:
		return "capture"
	case grokIRGroupNonCapture:
		return "noncapture"
	default:
		return "unknown"
	}
}

func formatRepeatMax(max int) string {
	if max < 0 {
		return "inf"
	}
	return strconv.Itoa(max)
}

func commonSuffix(left, right string) string {
	if left == "" || right == "" {
		return ""
	}
	max := len(left)
	if len(right) < max {
		max = len(right)
	}
	i := 0
	for i < max && left[len(left)-1-i] == right[len(right)-1-i] {
		i++
	}
	if i == 0 {
		return ""
	}
	return left[len(left)-i:]
}

func reverseString(s string) string {
	if len(s) <= 1 {
		return s
	}
	b := []byte(s)
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return string(b)
}

func stripMatchIRBoundaryAnchors(node *matchIRNode) (*matchIRNode, bool, bool) {
	if node == nil {
		return &matchIRNode{kind: matchIREmpty}, false, false
	}
	if node.kind == matchIRRawRegex {
		return &matchIRNode{kind: matchIREmpty}, isStartAnchorRaw(node.raw), isEndAnchorRaw(node.raw)
	}
	if node.kind != matchIRSequence {
		return node, false, false
	}

	children := node.children
	anchoredStart := false
	anchoredEnd := false

	for len(children) > 0 && children[0] != nil && children[0].kind == matchIRRawRegex && isStartAnchorRaw(children[0].raw) {
		anchoredStart = true
		children = children[1:]
	}
	for len(children) > 0 {
		last := children[len(children)-1]
		if last == nil || last.kind != matchIRRawRegex || !isEndAnchorRaw(last.raw) {
			break
		}
		anchoredEnd = true
		children = children[:len(children)-1]
	}

	return normalizeMatchSequence(children), anchoredStart, anchoredEnd
}

func isStartAnchorRaw(raw string) bool {
	return raw == "^" || raw == `\A`
}

func isEndAnchorRaw(raw string) bool {
	return raw == "$" || raw == `\z` || raw == `\Z`
}

func commonPrefix(a, b string) string {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	i := 0
	for i < n && a[i] == b[i] {
		i++
	}
	return a[:i]
}

func literalSet(lit string) map[string]struct{} {
	if lit == "" {
		return map[string]struct{}{}
	}
	return map[string]struct{}{lit: {}}
}

func copyLiteralSet(src map[string]struct{}) map[string]struct{} {
	dst := make(map[string]struct{}, len(src))
	for lit := range src {
		dst[lit] = struct{}{}
	}
	return dst
}

func intersectLiteralSets(a, b map[string]struct{}) map[string]struct{} {
	if len(a) == 0 || len(b) == 0 {
		return map[string]struct{}{}
	}
	out := make(map[string]struct{})
	if len(a) > len(b) {
		a, b = b, a
	}
	for lit := range a {
		if _, ok := b[lit]; ok {
			out[lit] = struct{}{}
		}
	}
	return out
}

func unionLiteralSets(dst map[string]struct{}, src map[string]struct{}) {
	for lit := range src {
		dst[lit] = struct{}{}
	}
}

func isLiteralByte(b byte) bool {
	switch b {
	case '(', ')', '[', ']', '|', '?', '*', '+', '{', '}', '^', '$', '.', '\\':
		return false
	case '%':
		return true
	default:
		return true
	}
}

func isEscapedLiteralByte(b byte) bool {
	switch b {
	case '\\', '.', '+', '*', '?', '(', ')', '[', ']', '{', '}', '|', '^', '$', '"', '\'', ':', '/', '-', '_', '%':
		return true
	default:
		return false
	}
}
