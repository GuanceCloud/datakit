// Package grok used to parses grok patterns in Go
package grok

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"unicode"
	"unicode/utf8"
)

// Denormalized patterns as regular expressions.

type SubMatchName struct {
	name []string

	subexpIndex []int
	subexpCount int
}

type GrokRegexp struct {
	grokPattern      *GrokPattern
	re               *regexp.Regexp
	filter           *regexpFilter
	subMatchNames    SubMatchName
	nameIndex        map[string]int
	valueKinds       []valueKind
	fastMatcher      *structuredMatcher
	prefilter        *regexpPrefilter
	multiFilter      multiPatternFilter
	requiredPrefix   string
	requiredSuffix   string
	requiredLiterals []string
	minMatchLength   int
	patternHash      uint64

	fastPathDisabled uint32
}

type RunPath uint8

const (
	RunPathRegexp RunPath = iota + 1
	RunPathFastPath
	RunPathFallback
)

func (p RunPath) String() string {
	switch p {
	case RunPathRegexp:
		return "regexp"
	case RunPathFastPath:
		return "fast_path"
	case RunPathFallback:
		return "fallback"
	default:
		return ""
	}
}

type FallbackReason uint8

const (
	FallbackNone FallbackReason = iota
	FallbackCompileRisk
	FallbackBudgetExceeded
	FallbackFastPathDisabled
	FallbackFastPathMismatch
)

func (r FallbackReason) String() string {
	switch r {
	case FallbackNone:
		return "none"
	case FallbackCompileRisk:
		return "compile_risk"
	case FallbackBudgetExceeded:
		return "budget_exceeded"
	case FallbackFastPathDisabled:
		return "fast_path_disabled"
	case FallbackFastPathMismatch:
		return "fast_path_mismatch"
	default:
		return ""
	}
}

type RunMeta struct {
	Path           RunPath
	FallbackReason FallbackReason
	PatternHash    uint64
	WorkUnits      int
}

type valueKind uint8

const (
	valueKindRaw valueKind = iota
	valueKindInt
	valueKindFloat
	valueKindBool
)

type stringBuffer struct {
	Values []string
}

type anyBuffer struct {
	Values []any
}

type stringBufferPool struct {
	size int
	pool sync.Pool
}

type anyBufferPool struct {
	size int
	pool sync.Pool
}

var ErrNotCompiled = errors.New("not compiled")
var ErrMismatch = errors.New("mismatch")

func (g *GrokRegexp) GetValByName(k string, val []string) (string, bool) {
	if len(val) != len(g.subMatchNames.name) {
		return "", false
	}
	if i, ok := g.nameIndex[k]; ok {
		return val[i], true
	}
	return "", false
}

func (g *GrokRegexp) MatchNames() []string {
	return g.subMatchNames.name
}

func (g *GrokRegexp) matchCount() int {
	return len(g.subMatchNames.name)
}

func (g *GrokRegexp) newStringBufferPool() *stringBufferPool {
	return newStringBufferPool(g.matchCount())
}

func (g *GrokRegexp) newAnyBufferPool() *anyBufferPool {
	return newAnyBufferPool(g.matchCount())
}

func (g *GrokRegexp) GetValAnyByName(k string, val []any) (any, bool) {
	if len(val) != len(g.subMatchNames.name) {
		return "", false
	}
	if i, ok := g.nameIndex[k]; ok {
		return val[i], true
	}
	return "", false
}

func (g *GrokRegexp) GetValCastByName(k string, val []string) (any, bool) {
	if len(val) != len(g.subMatchNames.name) {
		return nil, false
	}
	i, ok := g.nameIndex[k]
	if !ok {
		return nil, false
	}

	return castValue(val[i], g.valueKinds[i])
}

func (g *GrokRegexp) Run(content string, trimSpace bool) ([]string, error) {
	return g.runToWithMeta(content, trimSpace, nil, nil)
}

func (g *GrokRegexp) runTo(content string, trimSpace bool, dst []string) ([]string, error) {
	return g.runToWithMeta(content, trimSpace, dst, nil)
}

func (g *GrokRegexp) RunWithMeta(content string, trimSpace bool, meta *RunMeta) ([]string, error) {
	return g.runToWithMeta(content, trimSpace, nil, meta)
}

func (g *GrokRegexp) runToWithMeta(content string, trimSpace bool, dst []string, meta *RunMeta) ([]string, error) {
	g.initRunMeta(meta)
	if g.fastMatcher != nil {
		if g.fastPathIsDisabled() {
			setRunFallback(meta, RunPathRegexp, FallbackFastPathDisabled)
		} else {
			result := ensureStringBuffer(dst, len(g.subMatchNames.name))
			budget := newStructuredMatchBudget(len(content), g.fastMatcher.budgetSteps)
			if g.fastMatcher.match(result, content, trimSpace, &budget) {
				setRunMetaPath(meta, RunPathFastPath)
				setRunMetaWork(meta, budget.used)
				return result, nil
			}
			setRunMetaWork(meta, budget.used)
			if budget.exceeded {
				setRunFallback(meta, RunPathFallback, FallbackBudgetExceeded)
			} else {
				setRunFallback(meta, RunPathFallback, FallbackFastPathMismatch)
			}
		}
	}

	match, err := g.matchIndexes(content)
	if err != nil {
		return nil, err
	}

	result := ensureStringBuffer(dst, len(g.subMatchNames.name))

	for i := range g.subMatchNames.name {
		result[i] = extractMatch(content, match, g.subMatchNames.subexpIndex[i], trimSpace)
	}

	return result, nil
}

func (g *GrokRegexp) WithTypeInfo() bool {
	return len(g.grokPattern.varbType) > 0
}

func (g *GrokRegexp) RunWithTypeInfo(content string, trimSpace bool) ([]any, error) {
	return g.runWithTypeInfoToWithMeta(content, trimSpace, nil, nil)
}

func (g *GrokRegexp) runWithTypeInfoTo(content string, trimSpace bool, dst []any) ([]any, error) {
	return g.runWithTypeInfoToWithMeta(content, trimSpace, dst, nil)
}

func (g *GrokRegexp) RunWithTypeInfoWithMeta(content string, trimSpace bool, meta *RunMeta) ([]any, error) {
	return g.runWithTypeInfoToWithMeta(content, trimSpace, nil, meta)
}

func (g *GrokRegexp) runWithTypeInfoToWithMeta(content string, trimSpace bool, dst []any, meta *RunMeta) ([]any, error) {
	g.initRunMeta(meta)
	if g.fastMatcher != nil {
		if g.fastPathIsDisabled() {
			setRunFallback(meta, RunPathRegexp, FallbackFastPathDisabled)
		} else {
			castDst := ensureTypedBuffer(dst, g.valueKinds)
			budget := newStructuredMatchBudget(len(content), g.fastMatcher.budgetSteps)
			if g.fastMatcher.matchTyped(castDst, content, trimSpace, g.valueKinds, &budget) {
				setRunMetaPath(meta, RunPathFastPath)
				setRunMetaWork(meta, budget.used)
				return castDst, nil
			}
			setRunMetaWork(meta, budget.used)
			if budget.exceeded {
				setRunFallback(meta, RunPathFallback, FallbackBudgetExceeded)
			} else {
				setRunFallback(meta, RunPathFallback, FallbackFastPathMismatch)
			}
		}
	}

	match, err := g.matchIndexes(content)
	if err != nil {
		return nil, err
	}

	castDst := ensureTypedBuffer(dst, g.valueKinds)

	for i := range g.subMatchNames.name {
		raw := extractMatch(content, match, g.subMatchNames.subexpIndex[i], trimSpace)
		castDst[i], _ = castValue(raw, g.valueKinds[i])
	}

	return castDst, nil
}

func (g *GrokRegexp) PatternHash() uint64 {
	return g.patternHash
}

func (g *GrokRegexp) DisableFastPath() {
	atomic.StoreUint32(&g.fastPathDisabled, 1)
}

func (g *GrokRegexp) FastPathDisabled() bool {
	return g.fastPathIsDisabled()
}

func (g *GrokRegexp) fastPathIsDisabled() bool {
	return atomic.LoadUint32(&g.fastPathDisabled) != 0
}

func (g *GrokRegexp) initRunMeta(meta *RunMeta) {
	if meta == nil {
		return
	}
	*meta = RunMeta{
		Path:        RunPathRegexp,
		PatternHash: g.patternHash,
	}
}

func setRunMetaPath(meta *RunMeta, path RunPath) {
	if meta != nil {
		meta.Path = path
	}
}

func setRunMetaWork(meta *RunMeta, work int) {
	if meta != nil {
		meta.WorkUnits = work
	}
}

func setRunFallback(meta *RunMeta, path RunPath, reason FallbackReason) {
	if meta == nil {
		return
	}
	meta.Path = path
	meta.FallbackReason = reason
}

func (g *GrokRegexp) matchIndexes(content string) ([]int, error) {
	if g.re == nil {
		return nil, ErrNotCompiled
	}
	if g.multiFilter != nil && !g.multiFilter.MatchString(content) {
		return nil, ErrMismatch
	}
	if g.prefilter != nil && g.prefilter.rejects(content) {
		return nil, ErrMismatch
	}
	if g.filter != nil && !g.filter.re.MatchString(content) {
		return nil, ErrMismatch
	}

	match := g.re.FindStringSubmatchIndex(content)
	if len(match) == 0 {
		return nil, ErrMismatch
	}
	if g.subMatchNames.subexpCount*2 != len(match) {
		return nil, ErrMismatch
	}

	return match, nil
}

func extractMatch(content string, match []int, idx int, trimSpace bool) string {
	left := match[2*idx]
	right := match[2*idx+1]
	if left == -1 || right == -1 {
		return ""
	}

	if trimSpace {
		return trimMatch(content[left:right])
	}
	return content[left:right]
}

func ensureStringBuffer(dst []string, size int) []string {
	return clearStringBuffer(dst, size)
}

func ensureAnyBuffer(dst []any, size int) []any {
	return clearAnyBuffer(dst, size)
}

func ensureTypedBuffer(dst []any, kinds []valueKind) []any {
	dst = clearAnyBuffer(dst, len(kinds))
	for i := range dst {
		dst[i], _ = castValue("", kinds[i])
	}
	return dst
}

func newStringBufferPool(size int) *stringBufferPool {
	p := &stringBufferPool{size: size}
	p.pool.New = func() any {
		return &stringBuffer{Values: make([]string, 0, size)}
	}
	return p
}

func (p *stringBufferPool) get() *stringBuffer {
	return p.pool.Get().(*stringBuffer)
}

func (p *stringBufferPool) run(g *GrokRegexp, content string, trimSpace bool) (*stringBuffer, error) {
	buf := p.get()
	ret, err := g.runTo(content, trimSpace, buf.Values)
	if err != nil {
		p.put(buf)
		return nil, err
	}

	buf.Values = ret
	return buf, nil
}

func (p *stringBufferPool) put(buf *stringBuffer) {
	if buf == nil {
		return
	}

	buf.Values = clearStringBuffer(buf.Values, p.size)[:0]
	p.pool.Put(buf)
}

func newAnyBufferPool(size int) *anyBufferPool {
	p := &anyBufferPool{size: size}
	p.pool.New = func() any {
		return &anyBuffer{Values: make([]any, 0, size)}
	}
	return p
}

func (p *anyBufferPool) get() *anyBuffer {
	return p.pool.Get().(*anyBuffer)
}

func (p *anyBufferPool) run(g *GrokRegexp, content string, trimSpace bool) (*anyBuffer, error) {
	buf := p.get()
	ret, err := g.runWithTypeInfoTo(content, trimSpace, buf.Values)
	if err != nil {
		p.put(buf)
		return nil, err
	}

	buf.Values = ret
	return buf, nil
}

func (p *anyBufferPool) put(buf *anyBuffer) {
	if buf == nil {
		return
	}

	buf.Values = clearAnyBuffer(buf.Values, p.size)[:0]
	p.pool.Put(buf)
}

func clearStringBuffer(dst []string, size int) []string {
	if cap(dst) < size {
		return make([]string, size)
	}

	dst = dst[:size]
	for i := range dst {
		dst[i] = ""
	}
	return dst
}

func clearAnyBuffer(dst []any, size int) []any {
	if cap(dst) < size {
		return make([]any, size)
	}

	dst = dst[:size]
	for i := range dst {
		dst[i] = nil
	}
	return dst
}

func trimMatch(s string) string {
	if len(s) == 0 {
		return s
	}

	if !isASCIISpace(s[0]) && !isASCIISpace(s[len(s)-1]) {
		if s[0] < utf8.RuneSelf && s[len(s)-1] < utf8.RuneSelf {
			return s
		}

		first, _ := utf8.DecodeRuneInString(s)
		last, _ := utf8.DecodeLastRuneInString(s)
		if !unicode.IsSpace(first) && !unicode.IsSpace(last) {
			return s
		}
	}

	return strings.TrimSpace(s)
}

func isASCIISpace(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '\f', '\v':
		return true
	default:
		return false
	}
}

func castValue(raw string, kind valueKind) (any, bool) {
	switch kind {
	case valueKindInt:
		v, _ := strconv.ParseInt(raw, 0, 64)
		return v, true
	case valueKindFloat:
		v, _ := strconv.ParseFloat(raw, 64)
		return v, true
	case valueKindBool:
		v, _ := strconv.ParseBool(raw)
		return v, true
	case valueKindRaw:
		return raw, true
	default:
		return nil, false
	}
}
