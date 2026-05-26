package grok

import (
	internalmatch "github.com/GuanceCloud/grok/internal/match"
	"net"
	"strings"
	"sync"
	"unicode/utf8"
)

type structuredMatcher struct {
	steps                []structuredStep
	ir                   structuredIRInfo
	required             []string
	anchoredStart        bool
	anchoredEnd          bool
	anchoredRunner       *anchoredDissectRunner
	accessRunner         *accessLogRunner
	commonRunner         *commonApacheRunner
	jenkinsRunner        *jenkinsRunner
	tomcatRunner         *tomcatCatalinaRunner
	mysqlRunner          *mysqlSimpleRunner
	sqlserverRunner      *sqlServerRunner
	kafkaRunner          *kafkaBracketRunner
	kingbaseRunner       *kingbaseRunner
	redisRunner          *redisRunner
	damengRunner         *damengRunner
	elasticRunner        *elasticRunner
	elasticDefaultRunner *elasticDefaultRunner
	elasticSearchRunner  *elasticSearchSlowRunner
	nginxErrorRunner     *nginxErrorRunner
	postfixRunner        *postfixQueueRunner
	rabbitRunner         *rabbitMQRunner
	solrRunner           *solrRunner
	startOnlyRunner      bool
	writes               bool
	backtracking         bool
	changeLog            bool
	changeCap            int
	riskySearch          bool
}

type matchChange struct {
	idx  int
	prev string
}

type typedMatchChange struct {
	idx  int
	prev any
}

type matchChangeBuffer struct {
	changes []matchChange
}

type typedMatchChangeBuffer struct {
	changes []typedMatchChange
}

var matchChangePool = sync.Pool{
	New: func() any {
		return &matchChangeBuffer{changes: make([]matchChange, 0, 16)}
	},
}

var typedMatchChangePool = sync.Pool{
	New: func() any {
		return &typedMatchChangeBuffer{changes: make([]typedMatchChange, 0, 16)}
	},
}

type structuredStep struct {
	literal               string
	parser                *structuredParser
	submatcher            *structuredMatcher
	alternatives          []*structuredMatcher
	altPrefixes           []string
	captureIndex          int
	optional              bool
	optPrefix             string
	optPrefixSkips        bool
	deterministicOptional bool
	parserBacktracking    bool
	tailMinWidth          int
	ir                    structuredIRInfo
	nextIR                structuredIRInfo
	writes                bool
}

type structuredIRInfo = internalmatch.StructuredIRInfo

type structuredParser struct {
	alias         string
	dstIndex      int
	kind          structuredKind
	inner         *structuredParser
	charClass     *asciiCharClass
	allowEmpty    bool
	dotAll        bool
	singleChar    bool
	nextLiteral   string
	nextParser    structuredKind
	hasNextParser bool
	wrapPrefix    string
	wrapSuffix    string
}

type asciiCharClass = internalmatch.ASCIICharClass

type structuredKind = internalmatch.StructuredKind

const structuredMaxExpandedRepeat = 16

const (
	structuredWord               = internalmatch.StructuredWord
	structuredNotSpace           = internalmatch.StructuredNotSpace
	structuredHostName           = internalmatch.StructuredHostName
	structuredIPOrHost           = internalmatch.StructuredIPOrHost
	structuredNumber             = internalmatch.StructuredNumber
	structuredInt                = internalmatch.StructuredInt
	structuredPosInt             = internalmatch.StructuredPosInt
	structuredNonNegInt          = internalmatch.StructuredNonNegInt
	structuredCharClass          = internalmatch.StructuredCharClass
	structuredMonthName          = internalmatch.StructuredMonthName
	structuredDayName            = internalmatch.StructuredDayName
	structuredTimeOfDay          = internalmatch.StructuredTimeOfDay
	structuredYear               = internalmatch.StructuredYear
	structuredMonthNum           = internalmatch.StructuredMonthNum
	structuredMonthDay           = internalmatch.StructuredMonthDay
	structuredHour               = internalmatch.StructuredHour
	structuredMinute             = internalmatch.StructuredMinute
	structuredSecond             = internalmatch.StructuredSecond
	structuredQuoted             = internalmatch.StructuredQuoted
	structuredUntilLiteral       = internalmatch.StructuredUntilLiteral
	structuredGreedyUntilLiteral = internalmatch.StructuredGreedyUntilLiteral
	structuredSpaceOne           = internalmatch.StructuredSpaceOne
	structuredSpacePlus          = internalmatch.StructuredSpacePlus
	structuredSpaceStar          = internalmatch.StructuredSpaceStar
	structuredTimestampISO8601   = internalmatch.StructuredTimestampISO8601
	structuredURIPath            = internalmatch.StructuredURIPath
	structuredURIPathParam       = internalmatch.StructuredURIPathParam
	structuredHTTPDate           = internalmatch.StructuredHTTPDate
	structuredLogLevel           = internalmatch.StructuredLogLevel
)

func buildFastMatcher(gP *GrokPattern, storage PatternStorageIface, meta *compiledRegexpMeta) *structuredMatcher {
	if storage == nil {
		return nil
	}
	return buildStructuredFastMatcher(gP.pattern, storage, meta)
}

func buildStructuredFastMatcher(pattern string, storage PatternStorageIface, meta *compiledRegexpMeta) *structuredMatcher {
	steps, ok := compileStructuredSteps(pattern, storage, meta, 0)
	if !ok || len(steps) == 0 {
		if matcher := buildPatternOnlyFastMatcher(pattern, storage, meta); matcher != nil {
			return matcher
		}
		return nil
	}
	steps = compactStructuredSteps(steps)

	configureStructuredSteps(steps)
	if !hasStructuredMatcherWork(steps) || !shouldUseStructuredMatcher(steps) {
		return nil
	}

	matcher := structuredMatcher{
		steps:         steps,
		ir:            buildStructuredMatcherIR(steps),
		required:      buildStructuredRequiredLiterals(steps, 4),
		anchoredStart: patternHasStartAnchor(pattern),
		anchoredEnd:   patternHasEndAnchor(pattern),
		writes:        matcherWrites(steps),
		backtracking:  matcherNeedsBacktracking(steps),
		changeLog:     matcherNeedsChangeLog(steps),
		changeCap:     matcherChangeCapacity(steps),
	}
	matcher.anchoredRunner, _ = compileAnchoredDissectRunner(steps, true)
	if !matcher.anchoredEnd {
		matcher.accessRunner, _ = compileAccessRunner(steps)
		matcher.commonRunner, _ = compileCommonApacheRunner(pattern, meta.nameIndex, storage)
		matcher.jenkinsRunner, _ = compileJenkinsRunner(pattern, meta.nameIndex, storage)
		matcher.tomcatRunner, _ = compileTomcatCatalinaRunner(pattern, meta.nameIndex, storage)
		matcher.mysqlRunner, _ = compileMySQLSimpleRunner(pattern, meta.nameIndex, storage)
		matcher.sqlserverRunner, _ = compileSQLServerRunner(pattern, meta.nameIndex, storage)
		matcher.kafkaRunner, _ = compileKafkaBracketRunner(pattern, meta.nameIndex, storage)
		matcher.kingbaseRunner, _ = compileKingbaseRunner(pattern, meta.nameIndex, storage)
		matcher.redisRunner, _ = compileRedisRunner(pattern, meta.nameIndex, storage)
		matcher.damengRunner, _ = compileDamengRunner(pattern, meta.nameIndex, storage)
		matcher.elasticRunner, _ = compileElasticRunner(pattern, meta.nameIndex, storage)
		matcher.elasticDefaultRunner, _ = compileElasticDefaultRunner(pattern, meta.nameIndex, storage)
		matcher.elasticSearchRunner, _ = compileElasticSearchSlowRunner(pattern, meta.nameIndex, storage)
		matcher.nginxErrorRunner, _ = compileNginxErrorRunner(pattern, meta.nameIndex, storage)
		matcher.postfixRunner, _ = compilePostfixQueueRunner(pattern, meta.nameIndex, storage)
		matcher.rabbitRunner, _ = compileRabbitMQRunner(pattern, steps, storage)
		matcher.solrRunner, _ = compileSolrRunner(pattern, steps, storage)
	}
	matcher.startOnlyRunner = matcher.hasConcreteStartOnlyRunner()
	matcher.riskySearch = shouldLimitUnanchoredSearchMatcher(matcher)
	if matcher.riskySearch && shouldDisableSmallRiskySearchMatcher(matcher) {
		return nil
	}
	return &matcher
}

func buildPatternOnlyFastMatcher(pattern string, storage PatternStorageIface, meta *compiledRegexpMeta) *structuredMatcher {
	if meta == nil {
		return nil
	}
	matcher := structuredMatcher{
		anchoredStart: patternHasStartAnchor(pattern),
		anchoredEnd:   patternHasEndAnchor(pattern),
	}
	if !matcher.anchoredEnd {
		matcher.postfixRunner, _ = compilePostfixQueueRunner(pattern, meta.nameIndex, storage)
	}
	matcher.startOnlyRunner = matcher.hasConcreteStartOnlyRunner()
	if !matcher.startOnlyRunner {
		return nil
	}
	return &matcher
}

func getMatchChangeBuffer(capHint int) *matchChangeBuffer {
	buf := matchChangePool.Get().(*matchChangeBuffer)
	if capHint < 16 {
		capHint = 16
	}
	if cap(buf.changes) < capHint {
		buf.changes = make([]matchChange, 0, capHint)
	} else {
		buf.changes = buf.changes[:0]
	}
	return buf
}

func putMatchChangeBuffer(buf *matchChangeBuffer, changes []matchChange) {
	if cap(changes) > 256 {
		buf.changes = make([]matchChange, 0, 16)
	} else {
		buf.changes = changes[:0]
	}
	matchChangePool.Put(buf)
}

func getTypedMatchChangeBuffer(capHint int) *typedMatchChangeBuffer {
	buf := typedMatchChangePool.Get().(*typedMatchChangeBuffer)
	if capHint < 16 {
		capHint = 16
	}
	if cap(buf.changes) < capHint {
		buf.changes = make([]typedMatchChange, 0, capHint)
	} else {
		buf.changes = buf.changes[:0]
	}
	return buf
}

func putTypedMatchChangeBuffer(buf *typedMatchChangeBuffer, changes []typedMatchChange) {
	if cap(changes) > 256 {
		buf.changes = make([]typedMatchChange, 0, 16)
	} else {
		buf.changes = changes[:0]
	}
	typedMatchChangePool.Put(buf)
}

func compactStructuredSteps(steps []structuredStep) []structuredStep {
	if len(steps) == 0 {
		return steps
	}

	for i := range steps {
		if steps[i].submatcher != nil {
			steps[i].submatcher.steps = compactStructuredSteps(steps[i].submatcher.steps)
		}
		for _, alt := range steps[i].alternatives {
			alt.steps = compactStructuredSteps(alt.steps)
		}
	}

	out := make([]structuredStep, 0, len(steps))
	for i := 0; i < len(steps); i++ {
		if i+2 < len(steps) {
			if wrapped, ok := buildWrappedParserStep(steps[i], steps[i+1], steps[i+2], false); ok {
				out = append(out, wrapped)
				i += 2
				continue
			}
		}
		out = append(out, steps[i])
	}
	return out
}

func patternHasStartAnchor(pattern string) bool {
	return strings.HasPrefix(pattern, "^")
}

func patternHasEndAnchor(pattern string) bool {
	if !strings.HasSuffix(pattern, "$") {
		return false
	}
	backslashes := 0
	for i := len(pattern) - 2; i >= 0 && pattern[i] == '\\'; i-- {
		backslashes++
	}
	return backslashes%2 == 0
}

func configureStructuredSteps(steps []structuredStep) {
	for i := range steps {
		if steps[i].submatcher != nil {
			configureStructuredSteps(steps[i].submatcher.steps)
			steps[i].submatcher.ir = buildStructuredMatcherIR(steps[i].submatcher.steps)
			steps[i].submatcher.required = buildStructuredRequiredLiterals(steps[i].submatcher.steps, 4)
			steps[i].submatcher.writes = matcherWrites(steps[i].submatcher.steps)
			steps[i].submatcher.backtracking = matcherNeedsBacktracking(steps[i].submatcher.steps)
			steps[i].submatcher.changeLog = matcherNeedsChangeLog(steps[i].submatcher.steps)
			steps[i].submatcher.changeCap = matcherChangeCapacity(steps[i].submatcher.steps)
			steps[i].writes = steps[i].writes || steps[i].submatcher.writes
		}
		for _, alt := range steps[i].alternatives {
			configureStructuredSteps(alt.steps)
			alt.ir = buildStructuredMatcherIR(alt.steps)
			alt.required = buildStructuredRequiredLiterals(alt.steps, 4)
			alt.writes = matcherWrites(alt.steps)
			alt.backtracking = matcherNeedsBacktracking(alt.steps)
			alt.changeLog = matcherNeedsChangeLog(alt.steps)
			alt.changeCap = matcherChangeCapacity(alt.steps)
			steps[i].writes = steps[i].writes || alt.writes
		}
		if steps[i].captureIndex >= 0 {
			steps[i].writes = true
		}
		if steps[i].parser != nil {
			steps[i].writes = steps[i].writes || steps[i].parser.dstIndex >= 0
		}
	}

	tailMinWidth := 0
	nextIR := internalmatch.EmptyIRInfo()
	for i := len(steps) - 1; i >= 0; i-- {
		steps[i].ir = buildStructuredStepIR(steps[i])
		steps[i].nextIR = nextIR
		requiredMinWidth := steps[i].ir.MinWidth
		if steps[i].optional {
			requiredMinWidth = 0
		}
		steps[i].tailMinWidth = tailMinWidth + requiredMinWidth
		tailMinWidth = steps[i].tailMinWidth
		nextIR = internalmatch.CombineIR(steps[i].ir, nextIR)
	}

	for i := range steps {
		if steps[i].parser != nil {
			steps[i].parser.nextLiteral = nextStructuredSlicingLiteral(steps, i+1)
			steps[i].parser.nextParser, steps[i].parser.hasNextParser = nextStructuredParserKind(steps, i+1)
		}
		if len(steps[i].alternatives) > 0 {
			steps[i].altPrefixes = buildAlternativePrefixes(steps[i].alternatives)
		}
		if steps[i].optional {
			if steps[i].parser != nil && steps[i].parser.nextLiteral != "" && !steps[i].parser.allowEmpty {
				steps[i].optPrefix = steps[i].parser.nextLiteral
				steps[i].optPrefixSkips = true
			} else if steps[i].ir.FirstLiteralExact {
				steps[i].optPrefix = steps[i].ir.FirstLiteral
				steps[i].optPrefixSkips = false
			}
			steps[i].deterministicOptional = isDeterministicOptionalStep(steps, i)
		}
	}

	for i := range steps {
		steps[i].parserBacktracking = parserStepNeedsBacktracking(steps, i)
	}
}

func buildStructuredMatcherIR(steps []structuredStep) structuredIRInfo {
	info := internalmatch.EmptyIRInfo()
	if len(steps) == 0 {
		return info
	}

	for i := range steps {
		stepInfo := steps[i].ir
		info.MinWidth += stepInfo.MinWidth
		if !stepInfo.Nullable {
			info.Nullable = false
		}
	}

	info.FirstLiteral, info.FirstLiteralExact = matcherBoundaryLiteral(steps, true)
	info.LastLiteral, info.LastLiteralExact = matcherBoundaryLiteral(steps, false)
	return info
}

func buildStructuredStepIR(step structuredStep) structuredIRInfo {
	info := structuredIRInfo{}

	switch {
	case step.literal != "":
		info.MinWidth = len(step.literal)
		info.FirstLiteral = step.literal
		info.LastLiteral = step.literal
		info.FirstLiteralExact = true
		info.LastLiteralExact = true
	case step.parser != nil:
		info.MinWidth = parserMinWidth(step.parser)
		info.Nullable = step.parser.allowEmpty
		if step.parser.wrapPrefix != "" {
			info.FirstLiteral = step.parser.wrapPrefix
			info.FirstLiteralExact = true
		}
		if step.parser.wrapSuffix != "" {
			info.LastLiteral = step.parser.wrapSuffix
			info.LastLiteralExact = true
		}
	case step.submatcher != nil:
		info = step.submatcher.ir
	case len(step.alternatives) > 0:
		info = buildAlternativeIR(step.alternatives)
	default:
		info.Nullable = true
		info.FirstLiteralExact = true
		info.LastLiteralExact = true
	}

	if step.optional {
		info.Nullable = true
		info.MinWidth = 0
	}

	return info
}

func buildAlternativeIR(alts []*structuredMatcher) structuredIRInfo {
	info := structuredIRInfo{Nullable: false}
	if len(alts) == 0 {
		info.Nullable = true
		info.FirstLiteralExact = true
		info.LastLiteralExact = true
		return info
	}

	minWidth := -1
	firstLiteral := ""
	firstExact := true
	lastLiteral := ""
	lastExact := true
	nullable := false

	for i, alt := range alts {
		altInfo := alt.ir
		if i == 0 || altInfo.MinWidth < minWidth {
			minWidth = altInfo.MinWidth
		}
		nullable = nullable || altInfo.Nullable

		if i == 0 {
			firstLiteral = altInfo.FirstLiteral
			firstExact = altInfo.FirstLiteralExact
			lastLiteral = altInfo.LastLiteral
			lastExact = altInfo.LastLiteralExact
			continue
		}
		if firstLiteral != altInfo.FirstLiteral || !altInfo.FirstLiteralExact {
			firstLiteral = ""
			firstExact = false
		}
		if lastLiteral != altInfo.LastLiteral || !altInfo.LastLiteralExact {
			lastLiteral = ""
			lastExact = false
		}
	}

	info.MinWidth = minWidth
	info.Nullable = nullable
	info.FirstLiteral = firstLiteral
	info.FirstLiteralExact = firstExact && firstLiteral != ""
	info.LastLiteral = lastLiteral
	info.LastLiteralExact = lastExact && lastLiteral != ""
	return info
}

func matcherBoundaryLiteral(steps []structuredStep, forward bool) (string, bool) {
	parts := make([]structuredIRInfo, 0, len(steps))
	for i := range steps {
		parts = append(parts, steps[i].ir)
	}
	return internalmatch.IRBoundaryLiteral(parts, forward)
}

func parserMinWidth(p *structuredParser) int {
	if p == nil {
		return 0
	}
	if p.inner != nil {
		return len(p.wrapPrefix) + len(p.wrapSuffix) + parserMinWidth(p.inner)
	}
	switch p.kind {
	case structuredWord, structuredNotSpace, structuredHostName, structuredNumber, structuredInt, structuredPosInt, structuredNonNegInt, structuredCharClass:
		if p.allowEmpty {
			return 0
		}
		return 1
	case structuredMonthName:
		return 2
	case structuredDayName:
		return 3
	case structuredTimeOfDay:
		return 4
	case structuredYear:
		return 2
	case structuredMonthNum, structuredMonthDay, structuredHour, structuredMinute, structuredSecond:
		return 1
	case structuredQuoted:
		return 2
	case structuredUntilLiteral, structuredGreedyUntilLiteral, structuredSpaceStar:
		if p.singleChar {
			return 1
		}
		return 0
	case structuredSpaceOne, structuredSpacePlus:
		return 1
	case structuredTimestampISO8601:
		return 13
	case structuredHTTPDate:
		return 20
	case structuredLogLevel:
		return 2
	default:
		return 0
	}
}

func buildStructuredRequiredLiterals(steps []structuredStep, maxAtoms int) []string {
	chunks, ok := structuredRequiredLiteralChunks(steps, 32)
	if !ok || len(chunks) == 0 {
		return nil
	}
	return selectRequiredLiteralAtoms(chunks, maxAtoms)
}

func structuredRequiredLiteralChunks(steps []structuredStep, limit int) ([]string, bool) {
	if limit <= 0 {
		return nil, false
	}

	out := make([]string, 0, len(steps))
	for _, step := range steps {
		parts, ok := structuredRequiredLiteralChunksFromStep(step, limit-len(out))
		if !ok {
			return nil, false
		}
		out = append(out, parts...)
		if len(out) > limit {
			return nil, false
		}
	}
	return out, true
}

func structuredRequiredLiteralChunksFromStep(step structuredStep, limit int) ([]string, bool) {
	if step.optional || limit <= 0 {
		return nil, true
	}

	switch {
	case step.literal != "":
		return []string{step.literal}, true
	case step.submatcher != nil:
		return structuredRequiredLiteralChunks(step.submatcher.steps, limit)
	case len(step.alternatives) > 0:
		chunks, ok := structuredRequiredLiteralChunks(step.alternatives[0].steps, limit)
		if !ok {
			return nil, false
		}
		for i := 1; i < len(step.alternatives); i++ {
			other, ok := structuredRequiredLiteralChunks(step.alternatives[i].steps, limit)
			if !ok {
				return nil, false
			}
			chunks = commonLiteralSequence(chunks, other)
			if len(chunks) == 0 {
				return nil, true
			}
		}
		return chunks, true
	default:
		return nil, true
	}
}

func buildAlternativePrefixes(alts []*structuredMatcher) []string {
	if len(alts) < 2 {
		return nil
	}

	prefixes := make([]string, len(alts))
	for i, alt := range alts {
		if !alt.ir.FirstLiteralExact || alt.ir.FirstLiteral == "" {
			return nil
		}
		prefixes[i] = alt.ir.FirstLiteral
	}

	for i := range prefixes {
		for j := i + 1; j < len(prefixes); j++ {
			if prefixes[i] == prefixes[j] ||
				strings.HasPrefix(prefixes[i], prefixes[j]) ||
				strings.HasPrefix(prefixes[j], prefixes[i]) {
				return nil
			}
		}
	}

	return prefixes
}

func isDeterministicOptionalStep(steps []structuredStep, idx int) bool {
	step := steps[idx]
	if !step.optional || len(step.alternatives) > 0 {
		return false
	}
	if step.parser != nil {
		if step.parser.wrapPrefix != "" {
			curPrefix, ok := firstStructuredLiteral(step)
			if !ok || curPrefix == "" {
				return false
			}
			nextPrefix := ""
			if steps[idx].nextIR.FirstLiteralExact {
				nextPrefix = steps[idx].nextIR.FirstLiteral
			}
			if nextPrefix == "" {
				return false
			}
			return curPrefix != nextPrefix &&
				!strings.HasPrefix(curPrefix, nextPrefix) &&
				!strings.HasPrefix(nextPrefix, curPrefix)
		}
		return (step.parser.nextLiteral != "" && !step.parser.allowEmpty) ||
			canSkipOptionalParser(step.parser)
	}
	if step.submatcher != nil && step.submatcher.backtracking {
		return false
	}

	curPrefix, ok := firstStructuredLiteral(step)
	if !ok || curPrefix == "" {
		return false
	}
	nextPrefix := ""
	if steps[idx].nextIR.FirstLiteralExact {
		nextPrefix = steps[idx].nextIR.FirstLiteral
	}
	if nextPrefix == "" {
		return false
	}
	return curPrefix != nextPrefix &&
		!strings.HasPrefix(curPrefix, nextPrefix) &&
		!strings.HasPrefix(nextPrefix, curPrefix)
}

func canSkipOptionalParser(p *structuredParser) bool {
	if p == nil {
		return false
	}
	if p.nextLiteral != "" && !p.allowEmpty {
		return true
	}
	return p.dstIndex < 0 && p.kind == structuredYear && p.hasNextParser && p.nextParser == structuredTimeOfDay
}

func matcherWrites(steps []structuredStep) bool {
	for i := range steps {
		if steps[i].writes {
			return true
		}
	}
	return false
}

func matcherNeedsBacktracking(steps []structuredStep) bool {
	for i := range steps {
		if steps[i].parserBacktracking {
			return true
		}
		if steps[i].optional && !isDeterministicOptionalStep(steps, i) && !optionalStepCanCommitLinearly(steps, i) {
			return true
		}
		if steps[i].submatcher != nil && matcherNeedsBacktracking(steps[i].submatcher.steps) {
			return true
		}
		for _, alt := range steps[i].alternatives {
			if matcherNeedsBacktracking(alt.steps) {
				return true
			}
		}
	}
	return false
}

func matcherNeedsChangeLog(steps []structuredStep) bool {
	for i := range steps {
		step := steps[i]
		if step.parserBacktracking && step.parser.dstIndex >= 0 {
			return true
		}
		if step.optional && !isDeterministicOptionalStep(steps, i) && !optionalStepCanCommitLinearly(steps, i) {
			return true
		}
		if step.submatcher != nil && matcherNeedsChangeLog(step.submatcher.steps) {
			return true
		}
		if len(step.alternatives) > 0 {
			if step.captureIndex >= 0 {
				return true
			}
			for _, alt := range step.alternatives {
				if alt.writes || matcherNeedsChangeLog(alt.steps) {
					return true
				}
			}
		}
	}
	return false
}

func parserNeedsBacktracking(step structuredStep) bool {
	return step.parser != nil && step.parser.nextLiteral != "" &&
		(step.parser.kind == structuredGreedyUntilLiteral ||
			(step.parser.kind == structuredNotSpace && !step.parser.hasNextParser && !literalStartsWithRegexpSpace(step.parser.nextLiteral)))
}

func parserStepNeedsBacktracking(steps []structuredStep, idx int) bool {
	if idx < 0 || idx >= len(steps) || !parserNeedsBacktracking(steps[idx]) {
		return false
	}
	step := steps[idx]
	if step.optional && step.deterministicOptional && step.parser != nil && step.parser.wrapSuffix != "" {
		return false
	}
	return !backtrackingParserCanCommitLinearly(steps, idx)
}

func backtrackingParserCanCommitLinearly(steps []structuredStep, idx int) bool {
	if idx < 0 || idx+1 >= len(steps) {
		return false
	}
	step := steps[idx]
	if step.optional || step.parser == nil || step.parser.nextLiteral == "" {
		return false
	}
	if idx+2 != len(steps) {
		return false
	}
	next := steps[idx+1]
	return !next.optional &&
		next.literal == step.parser.nextLiteral &&
		next.captureIndex < 0 &&
		next.parser == nil &&
		next.submatcher == nil &&
		len(next.alternatives) == 0
}

func optionalStepCanCommitLinearly(steps []structuredStep, idx int) bool {
	step := steps[idx]
	if !step.optional || step.parser == nil {
		return false
	}
	if step.captureIndex >= 0 || step.submatcher != nil || len(step.alternatives) > 0 {
		return false
	}
	if idx+1 >= len(steps) {
		return true
	}
	for i := idx + 1; i < len(steps); i++ {
		next := steps[i]
		if next.optional || next.captureIndex >= 0 || next.submatcher != nil || len(next.alternatives) > 0 || next.literal != "" || next.parser == nil {
			return false
		}
		if next.parser.dstIndex >= 0 || next.parser.kind != structuredGreedyUntilLiteral || next.parser.nextLiteral != "" {
			return false
		}
	}
	return true
}

func literalStartsWithRegexpSpace(lit string) bool {
	return lit != "" && isRegexpASCIISpace(lit[0])
}

func matcherChangeCapacity(steps []structuredStep) int {
	total := 0
	for i := range steps {
		total += stepChangeCapacity(steps[i])
	}
	return total
}

func stepChangeCapacity(step structuredStep) int {
	total := 0
	if step.parser != nil && step.parser.dstIndex >= 0 {
		total++
	}
	if step.captureIndex >= 0 {
		total++
	}
	if step.submatcher != nil {
		total += matcherChangeCapacity(step.submatcher.steps)
	}
	if len(step.alternatives) > 0 {
		maxAlt := 0
		for _, alt := range step.alternatives {
			altWrites := matcherChangeCapacity(alt.steps)
			if altWrites > maxAlt {
				maxAlt = altWrites
			}
		}
		total += maxAlt
	}
	return total
}

func hasStructuredMatcherWork(steps []structuredStep) bool {
	for i := range steps {
		switch {
		case steps[i].parser != nil:
			return true
		case steps[i].submatcher != nil && hasStructuredMatcherWork(steps[i].submatcher.steps):
			return true
		case len(steps[i].alternatives) > 0:
			for _, alt := range steps[i].alternatives {
				if hasStructuredMatcherWork(alt.steps) {
					return true
				}
			}
		}
	}
	return false
}

func shouldUseStructuredMatcher(steps []structuredStep) bool {
	for i := range steps {
		if steps[i].parser == nil {
			continue
		}
		// Unbounded leading GREEDYDATA is still risky, but a bounded one with a
		// concrete following literal is fine and now benchmarks well on real
		// access-log fixtures.
		if i == 0 && steps[i].parser.kind == structuredGreedyUntilLiteral && steps[i].parser.nextLiteral == "" {
			return false
		}
		break
	}

	stats := collectStructuredStats(steps)
	if stats.newlineLiteralCount > 0 && stats.alternativeCount > 0 && stats.parserCount > 12 {
		return false
	}
	return true
}

func shouldLimitUnanchoredSearchMatcher(matcher structuredMatcher) bool {
	if matcher.anchoredStart || matcher.hasStartOnlyRunner() {
		return false
	}
	if matcher.ir.FirstLiteralExact && matcher.ir.FirstLiteral != "" {
		return false
	}
	return matcher.backtracking && matcher.changeLog && hasWritableOptionalStep(matcher.steps)
}

func hasWritableOptionalStep(steps []structuredStep) bool {
	for i := range steps {
		step := steps[i]
		if step.optional && step.writes {
			return true
		}
		if step.submatcher != nil && hasWritableOptionalStep(step.submatcher.steps) {
			return true
		}
		for _, alt := range step.alternatives {
			if hasWritableOptionalStep(alt.steps) {
				return true
			}
		}
	}
	return false
}

func shouldDisableSmallRiskySearchMatcher(matcher structuredMatcher) bool {
	stats := collectStructuredStats(matcher.steps)
	return stats.parserCount <= 5 && stats.submatcherCount <= 1 && stats.alternativeCount == 0
}

type structuredStats struct {
	parserCount         int
	optionalCount       int
	alternativeCount    int
	submatcherCount     int
	greedyParserCount   int
	dataParserCount     int
	newlineLiteralCount int
}

func collectStructuredStats(steps []structuredStep) structuredStats {
	var stats structuredStats
	for i := range steps {
		step := steps[i]
		if step.literal != "" && strings.Contains(step.literal, "\n") {
			stats.newlineLiteralCount++
		}
		if step.parser != nil {
			stats.parserCount++
			if step.parser.kind == structuredGreedyUntilLiteral {
				stats.greedyParserCount++
			}
			if step.parser.kind == structuredUntilLiteral || step.parser.kind == structuredGreedyUntilLiteral {
				stats.dataParserCount++
			}
		}
		if step.optional {
			stats.optionalCount++
		}
		if step.submatcher != nil {
			stats.submatcherCount++
			child := collectStructuredStats(step.submatcher.steps)
			stats.parserCount += child.parserCount
			stats.optionalCount += child.optionalCount
			stats.alternativeCount += child.alternativeCount
			stats.submatcherCount += child.submatcherCount
			stats.greedyParserCount += child.greedyParserCount
			stats.dataParserCount += child.dataParserCount
			stats.newlineLiteralCount += child.newlineLiteralCount
		}
		if len(step.alternatives) > 0 {
			stats.alternativeCount += len(step.alternatives)
			for _, alt := range step.alternatives {
				child := collectStructuredStats(alt.steps)
				stats.parserCount += child.parserCount
				stats.optionalCount += child.optionalCount
				stats.alternativeCount += child.alternativeCount
				stats.submatcherCount += child.submatcherCount
				stats.greedyParserCount += child.greedyParserCount
				stats.dataParserCount += child.dataParserCount
				stats.newlineLiteralCount += child.newlineLiteralCount
			}
		}
	}
	return stats
}

func compileStructuredSteps(pattern string, storage PatternStorageIface, meta *compiledRegexpMeta, depth int) ([]structuredStep, bool) {
	if depth > 16 {
		return nil, false
	}
	if step, ok := compileDotAllTailPattern(pattern); ok {
		return []structuredStep{step}, true
	}

	alts, pos, ok := parseStructuredExpression(pattern, 0, 0, storage, meta, depth)
	if !ok || pos != len(pattern) {
		return nil, false
	}

	if len(alts) == 1 {
		return alts[0], validateStructuredSteps(alts[0])
	}

	steps := []structuredStep{{alternatives: buildAlternativeMatchers(alts), captureIndex: -1}}
	return steps, validateStructuredSteps(steps)
}

func compileDotAllTailPattern(pattern string) (structuredStep, bool) {
	switch pattern {
	case `(?s)(.*)`, `(?s)(?:.*)`:
		return structuredStep{
			captureIndex: -1,
			parser: &structuredParser{
				dstIndex:   -1,
				kind:       structuredGreedyUntilLiteral,
				allowEmpty: true,
				dotAll:     true,
			},
		}, true
	case `(?s)(.+)`, `(?s)(?:.+)`:
		return structuredStep{
			captureIndex: -1,
			parser: &structuredParser{
				dstIndex:   -1,
				kind:       structuredGreedyUntilLiteral,
				allowEmpty: false,
				dotAll:     true,
			},
		}, true
	default:
		return structuredStep{}, false
	}
}

func parseStructuredExpression(pattern string, pos int, stop byte, storage PatternStorageIface, meta *compiledRegexpMeta, depth int) ([][]structuredStep, int, bool) {
	current := make([]structuredStep, 0, 8)
	alternatives := make([][]structuredStep, 0, 2)

	for pos < len(pattern) {
		if stop != 0 && pattern[pos] == stop {
			break
		}
		if pattern[pos] == '|' {
			alternatives = append(alternatives, current)
			current = make([]structuredStep, 0, 4)
			pos++
			continue
		}

		term, next, ok := parseStructuredTerm(pattern, pos, storage, meta, depth)
		if !ok {
			return nil, pos, false
		}
		current = append(current, term...)
		pos = next
	}

	alternatives = append(alternatives, current)
	return alternatives, pos, true
}

func parseStructuredTerm(pattern string, pos int, storage PatternStorageIface, meta *compiledRegexpMeta, depth int) ([]structuredStep, int, bool) {
	switch {
	case strings.HasPrefix(pattern[pos:], "%{"):
		return parseStructuredRef(pattern, pos, storage, meta, depth)
	case pattern[pos] == '(':
		return parseStructuredGroup(pattern, pos, storage, meta, depth)
	default:
		return parseStructuredLiteral(pattern, pos, depth)
	}
}

func parseStructuredRef(pattern string, pos int, storage PatternStorageIface, meta *compiledRegexpMeta, depth int) ([]structuredStep, int, bool) {
	end := strings.IndexByte(pattern[pos+2:], '}')
	if end < 0 {
		return nil, pos, false
	}
	end += pos + 2

	ref, ok, err := parsePatternRef(pattern[pos+2 : end])
	if err != nil || !ok {
		return nil, pos, false
	}
	min, max, next, ok := parseStructuredRepeat(pattern, end+1)
	if !ok {
		return nil, pos, false
	}

	if kind, primitive := structuredPrimitiveKind(ref.syntax); primitive && canUseStructuredPrimitive(ref.syntax, storage) {
		dstIndex := captureIndexForRef(ref, meta)
		if ref.alias != "" && dstIndex < 0 {
			return nil, pos, false
		}
		if !(min == 1 && max == 1) && !(min == 0 && max == 1) {
			return nil, pos, false
		}
		return []structuredStep{{
			captureIndex: -1,
			optional:     min == 0 && max == 1,
			parser: &structuredParser{
				alias:      ref.alias,
				dstIndex:   dstIndex,
				kind:       kind,
				allowEmpty: kind == structuredUntilLiteral || kind == structuredGreedyUntilLiteral,
				dotAll:     ref.syntax == "GREEDYLINES",
			},
		}}, next, true
	}

	child, exists := storage.GetPattern(ref.syntax)
	if !exists {
		return nil, pos, false
	}
	childSteps, ok := compileStructuredSteps(child.pattern, storage, meta, depth+1)
	if !ok {
		return nil, pos, false
	}

	if ref.alias == "" && ref.varType == "" && min == 1 && max == 1 {
		return childSteps, next, true
	}

	dstIndex := captureIndexForRef(ref, meta)
	if ref.alias != "" && dstIndex < 0 {
		return nil, pos, false
	}
	optional := min == 0 && max == 1
	if !optional && min == 1 && max == 1 {
		if inlined, ok := inlineAliasedStructuredSteps(childSteps, dstIndex); ok {
			return inlined, next, true
		}
	}
	repeated := expandStructuredRepeat(structuredStep{
		submatcher:   &structuredMatcher{steps: childSteps},
		captureIndex: dstIndex,
	}, min, max, false)
	if repeated == nil {
		return nil, pos, false
	}
	return repeated, next, true
}

func inlineAliasedStructuredSteps(steps []structuredStep, dstIndex int) ([]structuredStep, bool) {
	if dstIndex < 0 || len(steps) != 1 {
		return nil, false
	}

	step := steps[0]
	if step.optional || step.submatcher != nil || step.writes {
		return nil, false
	}

	switch {
	case step.parser != nil:
		parser := *step.parser
		parser.dstIndex = dstIndex
		step.parser = &parser
		return []structuredStep{step}, true
	case len(step.alternatives) > 0:
		step.captureIndex = dstIndex
		return []structuredStep{step}, true
	case step.literal != "":
		step.captureIndex = dstIndex
		return []structuredStep{step}, true
	default:
		return nil, false
	}
}

func parseStructuredGroup(pattern string, pos int, storage PatternStorageIface, meta *compiledRegexpMeta, depth int) ([]structuredStep, int, bool) {
	start := pos + 1
	if strings.HasPrefix(pattern[start:], "?:") {
		start += 2
	}

	alts, next, ok := parseStructuredExpression(pattern, start, ')', storage, meta, depth+1)
	if !ok || next >= len(pattern) || pattern[next] != ')' {
		return nil, pos, false
	}
	next++

	min, max, next, ok := parseStructuredRepeat(pattern, next)
	if !ok {
		return nil, pos, false
	}
	optional := min == 0 && max == 1

	if len(alts) == 1 && min == 1 && max == 1 {
		return alts[0], next, true
	}
	if len(alts) == 1 && optional {
		if flattened, ok := flattenOptionalStructuredGroup(alts[0]); ok {
			return flattened, next, true
		}
	}

	step := structuredStep{
		captureIndex: -1,
	}
	if len(alts) == 1 {
		step.submatcher = &structuredMatcher{steps: alts[0]}
	} else {
		step.alternatives = buildAlternativeMatchers(alts)
	}
	repeated := expandStructuredRepeat(step, min, max, true)
	if repeated == nil && max != 0 {
		return nil, pos, false
	}
	return repeated, next, true
}

func flattenOptionalStructuredGroup(steps []structuredStep) ([]structuredStep, bool) {
	if len(steps) != 1 {
		if wrapped, ok := flattenWrappedStructuredGroup(steps); ok {
			return []structuredStep{wrapped}, true
		}
		return nil, false
	}
	step := steps[0]
	if step.optional {
		return nil, false
	}
	step.optional = true
	return []structuredStep{step}, true
}

func flattenWrappedStructuredGroup(steps []structuredStep) (structuredStep, bool) {
	if len(steps) != 3 {
		return structuredStep{}, false
	}
	return buildWrappedParserStep(steps[0], steps[1], steps[2], true)
}

func buildWrappedParserStep(prefix, middle, suffix structuredStep, optional bool) (structuredStep, bool) {
	if prefix.literal == "" || suffix.literal == "" {
		return structuredStep{}, false
	}
	if prefix.optional || suffix.optional || prefix.captureIndex >= 0 || suffix.captureIndex >= 0 {
		return structuredStep{}, false
	}
	if middle.optional || middle.captureIndex >= 0 || middle.submatcher != nil || len(middle.alternatives) > 0 || middle.parser == nil {
		return structuredStep{}, false
	}
	if !canWrapStructuredParser(middle.parser) {
		return structuredStep{}, false
	}

	innerCopy := *middle.parser
	innerCopy.inner = nil
	innerCopy.wrapPrefix = ""
	innerCopy.wrapSuffix = ""
	innerCopy.nextLiteral = suffix.literal
	innerCopy.nextParser = 0
	innerCopy.hasNextParser = false
	parserCopy := innerCopy
	parserCopy.inner = &innerCopy
	parserCopy.wrapPrefix = prefix.literal
	parserCopy.wrapSuffix = suffix.literal

	return structuredStep{
		parser:       &parserCopy,
		captureIndex: -1,
		optional:     optional,
	}, true
}

func canWrapStructuredParser(p *structuredParser) bool {
	if p == nil || p.dstIndex < 0 || p.allowEmpty {
		return false
	}
	switch p.kind {
	case structuredWord, structuredNotSpace, structuredHostName, structuredIPOrHost,
		structuredNumber, structuredInt, structuredPosInt, structuredNonNegInt,
		structuredMonthName, structuredDayName, structuredTimeOfDay,
		structuredYear, structuredMonthNum, structuredMonthDay, structuredHour,
		structuredMinute, structuredSecond, structuredTimestampISO8601,
		structuredHTTPDate, structuredLogLevel, structuredURIPath, structuredURIPathParam:
		return true
	default:
		return false
	}
}

func parseStructuredLiteral(pattern string, pos int, depth int) ([]structuredStep, int, bool) {
	start := pos
	for pos < len(pattern) {
		switch pattern[pos] {
		case '\\':
			if pos+1 >= len(pattern) {
				return nil, start, false
			}
			pos += 2
		case '%':
			if strings.HasPrefix(pattern[pos:], "%{") {
				goto done
			}
			pos++
		case '(':
			goto done
		case ')', '|':
			goto done
		default:
			pos++
		}
	}

done:
	steps, ok := compileLiteralSteps(pattern[start:pos], depth == 0 && start == 0, depth == 0 && pos == len(pattern))
	return steps, pos, ok
}

func buildAlternativeMatchers(alts [][]structuredStep) []*structuredMatcher {
	out := make([]*structuredMatcher, 0, len(alts))
	for _, alt := range alts {
		out = append(out, &structuredMatcher{steps: alt})
	}
	return out
}

func captureIndexForRef(ref patternRef, meta *compiledRegexpMeta) int {
	if ref.alias == "" {
		return -1
	}
	if idx, ok := meta.nameIndex[ref.alias]; ok {
		return idx
	}
	return -1
}

func validateStructuredSteps(steps []structuredStep) bool {
	for i := range steps {
		step := steps[i]
		if step.parser != nil && i+1 < len(steps) && steps[i+1].parser != nil && !canParserEndWithoutLiteral(step.parser.kind) {
			if !canParserBridgeToFutureLiteral(steps, i) {
				return false
			}
		}
		if step.submatcher != nil && !validateStructuredSteps(step.submatcher.steps) {
			return false
		}
		for _, alt := range step.alternatives {
			if !validateStructuredSteps(alt.steps) {
				return false
			}
		}
	}
	return true
}

func canParserBridgeToFutureLiteral(steps []structuredStep, idx int) bool {
	if idx < 0 || idx >= len(steps) {
		return false
	}
	if steps[idx].parser == nil {
		return false
	}
	switch steps[idx].parser.kind {
	case structuredUntilLiteral, structuredGreedyUntilLiteral:
	default:
		return false
	}

	for i := idx + 1; i < len(steps); i++ {
		step := steps[i]
		if step.optional || step.submatcher != nil || len(step.alternatives) > 0 {
			return false
		}
		if step.literal != "" {
			return true
		}
		if step.parser == nil {
			return false
		}
		switch step.parser.kind {
		case structuredSpaceOne, structuredSpacePlus, structuredSpaceStar:
			continue
		default:
			return false
		}
	}

	return false
}

func structuredPrimitiveKind(syntax string) (structuredKind, bool) {
	switch syntax {
	case "WORD":
		return structuredWord, true
	case "HOSTNAME", "HOST":
		return structuredHostName, true
	case "IPORHOST":
		return structuredIPOrHost, true
	case "NOTSPACE":
		return structuredNotSpace, true
	case "URIPATH":
		return structuredURIPath, true
	case "URIPATHPARAM":
		return structuredURIPathParam, true
	case "INT":
		return structuredInt, true
	case "NUMBER", "BASE10NUM":
		return structuredNumber, true
	case "POSINT":
		return structuredPosInt, true
	case "NONNEGINT":
		return structuredNonNegInt, true
	case "MONTH":
		return structuredMonthName, true
	case "DAY":
		return structuredDayName, true
	case "TIME":
		return structuredTimeOfDay, true
	case "YEAR":
		return structuredYear, true
	case "MONTHNUM", "MONTHNUM2":
		return structuredMonthNum, true
	case "MONTHDAY":
		return structuredMonthDay, true
	case "HOUR":
		return structuredHour, true
	case "MINUTE":
		return structuredMinute, true
	case "SECOND":
		return structuredSecond, true
	case "QS", "QUOTEDSTRING":
		return structuredQuoted, true
	case "DATA":
		return structuredUntilLiteral, true
	case "GREEDYLINES":
		return structuredGreedyUntilLiteral, true
	case "GREEDYDATA":
		return structuredGreedyUntilLiteral, true
	case "SPACE":
		return structuredSpaceStar, true
	case "TIMESTAMP_ISO8601":
		return structuredTimestampISO8601, true
	case "HTTPDATE":
		return structuredHTTPDate, true
	case "LOGLEVEL":
		return structuredLogLevel, true
	default:
		return 0, false
	}
}

func canUseStructuredPrimitive(syntax string, storage PatternStorageIface) bool {
	if storage == nil {
		return true
	}
	gp, ok := storage.GetPattern(syntax)
	if !ok {
		return true
	}
	defaultPattern, hasDefault := defalutPatterns[syntax]
	if !hasDefault {
		return false
	}
	return gp.pattern == defaultPattern
}

func compileLiteralSteps(raw string, allowLeadingStartAnchor, allowTrailingEndAnchor bool) ([]structuredStep, bool) {
	if raw == "" {
		return nil, true
	}

	steps := make([]structuredStep, 0, 4)
	var b strings.Builder
	b.Grow(len(raw))
	flushLiteral := func() {
		if b.Len() > 0 {
			steps = append(steps, structuredStep{literal: b.String(), captureIndex: -1})
			b.Reset()
		}
	}
	appendOptionalLiteral := func(lit string) {
		flushLiteral()
		steps = append(steps, structuredStep{literal: lit, optional: true, captureIndex: -1})
	}
	appendRepeatedLiteral := func(ch byte, allowEmpty bool) {
		flushLiteral()
		class := &asciiCharClass{}
		class.Table[ch] = true
		steps = append(steps, structuredStep{
			captureIndex: -1,
			parser: &structuredParser{
				dstIndex:   -1,
				kind:       structuredCharClass,
				charClass:  class,
				allowEmpty: allowEmpty,
			},
		})
	}
	for i := 0; i < len(raw); i++ {
		switch raw[i] {
		case '\\':
			if i+1 >= len(raw) {
				return nil, false
			}
			i++
			if i+1 < len(raw) && raw[i+1] == '?' {
				switch raw[i] {
				case 't':
					appendOptionalLiteral("\t")
				case 'n':
					appendOptionalLiteral("\n")
				case 'r':
					appendOptionalLiteral("\r")
				case 's':
					return nil, false
				default:
					if (raw[i] >= 'A' && raw[i] <= 'Z') || (raw[i] >= 'a' && raw[i] <= 'z') {
						return nil, false
					}
					appendOptionalLiteral(string(raw[i]))
				}
				i++
				continue
			}
			switch raw[i] {
			case 's':
				flushLiteral()
				if i+1 < len(raw) && raw[i+1] == '?' {
					return nil, false
				}
				if i+1 >= len(raw) || (raw[i+1] != '+' && raw[i+1] != '*') {
					steps = append(steps, structuredStep{captureIndex: -1, parser: &structuredParser{dstIndex: -1, kind: structuredSpaceOne}})
					continue
				}
				if i+1 < len(raw) && raw[i+1] == '+' {
					steps = append(steps, structuredStep{captureIndex: -1, parser: &structuredParser{dstIndex: -1, kind: structuredSpacePlus}})
					i++
					continue
				}
				if i+1 < len(raw) && raw[i+1] == '*' {
					steps = append(steps, structuredStep{captureIndex: -1, parser: &structuredParser{dstIndex: -1, kind: structuredSpaceStar}})
					i++
					continue
				}
				b.WriteByte('s')
			case 't':
				b.WriteByte('\t')
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			default:
				if (raw[i] >= 'A' && raw[i] <= 'Z') || (raw[i] >= 'a' && raw[i] <= 'z') {
					return nil, false
				}
				b.WriteByte(raw[i])
			}
		case '^':
			if i != 0 || !allowLeadingStartAnchor {
				return nil, false
			}
			continue
		case '$':
			if i != len(raw)-1 || !allowTrailingEndAnchor {
				return nil, false
			}
			continue
		case '.':
			if i+1 < len(raw) && raw[i+1] == '*' {
				flushLiteral()
				steps = append(steps, structuredStep{captureIndex: -1, parser: &structuredParser{dstIndex: -1, kind: structuredGreedyUntilLiteral, allowEmpty: true}})
				i++
				continue
			}
			flushLiteral()
			steps = append(steps, structuredStep{
				captureIndex: -1,
				parser: &structuredParser{
					dstIndex:   -1,
					kind:       structuredUntilLiteral,
					singleChar: true,
				},
			})
		case '[':
			end := i + 1
			for end < len(raw) && raw[end] != ']' {
				if raw[end] == '\\' {
					end += 2
					continue
				}
				end++
			}
			if end >= len(raw) {
				return nil, false
			}
			classSteps, consumed, ok := compileSimpleCharClass(raw, i, end)
			if !ok {
				return nil, false
			}
			flushLiteral()
			steps = append(steps, classSteps...)
			i = consumed
			continue
		case ']', '{', '}', '+', '*', '?':
			return nil, false
		default:
			if i+1 < len(raw) && (raw[i+1] == '+' || raw[i+1] == '*') {
				appendRepeatedLiteral(raw[i], raw[i+1] == '*')
				i++
				continue
			}
			if i+1 < len(raw) && raw[i+1] == '?' {
				appendOptionalLiteral(string(raw[i]))
				i++
				continue
			}
			b.WriteByte(raw[i])
		}
	}

	flushLiteral()
	return steps, true
}

func compileSimpleCharClass(raw string, start, end int) ([]structuredStep, int, bool) {
	spec := raw[start+1 : end]
	if spec == "" {
		return nil, 0, false
	}

	class, ok := buildASCIICharClass(spec)
	if !ok {
		return nil, 0, false
	}

	if end+1 < len(raw) && (raw[end+1] == '*' || raw[end+1] == '+') {
		return []structuredStep{{
			captureIndex: -1,
			parser: &structuredParser{
				dstIndex:   -1,
				kind:       structuredCharClass,
				charClass:  class,
				allowEmpty: raw[end+1] == '*',
			},
		}}, end + 1, true
	}

	base := structuredStep{
		captureIndex: -1,
		parser: &structuredParser{
			dstIndex:   -1,
			kind:       structuredCharClass,
			charClass:  class,
			singleChar: true,
		},
	}

	return []structuredStep{base}, end, true
}

func parseStructuredRepeat(pattern string, pos int) (int, int, int, bool) {
	if pos >= len(pattern) {
		return 1, 1, pos, true
	}

	switch pattern[pos] {
	case '?':
		next := pos + 1
		if next < len(pattern) && pattern[next] == '?' {
			return 0, 0, 0, false
		}
		return 0, 1, next, true
	case '{':
		min, max, next, err := parseRepeatBounds(pattern, pos)
		if err != nil {
			return 0, 0, 0, false
		}
		if next < len(pattern) && pattern[next] == '?' {
			return 0, 0, 0, false
		}
		return min, max, next, true
	case '*', '+':
		return 0, 0, 0, false
	default:
		return 1, 1, pos, true
	}
}

func expandStructuredRepeat(base structuredStep, min, max int, allowInlineZero bool) []structuredStep {
	if min < 0 || max < min || max < 0 || max > structuredMaxExpandedRepeat {
		return nil
	}
	if max == 0 {
		if allowInlineZero {
			return []structuredStep{}
		}
		return nil
	}
	if min == 1 && max == 1 {
		return []structuredStep{base}
	}
	if min == 0 && max == 1 {
		base.optional = true
		return []structuredStep{base}
	}

	out := make([]structuredStep, 0, max)
	for i := 0; i < max; i++ {
		step := base
		step.optional = i >= min
		out = append(out, step)
	}
	return out
}

func buildASCIICharClass(spec string) (*asciiCharClass, bool) {
	negated := spec[0] == '^'
	if negated {
		spec = spec[1:]
		if spec == "" {
			return nil, false
		}
	}
	if strings.Contains(spec, "[:") || strings.Contains(spec, ":]") {
		return nil, false
	}

	class := &asciiCharClass{}
	for i := 0; i < len(spec); i++ {
		ch, next, special, ok := parseASCIICharClassAtom(spec, i)
		if !ok {
			return nil, false
		}
		if special != 0 {
			addASCIICharClassSpecial(class, special)
			i = next - 1
			continue
		}
		if next < len(spec) && spec[next] == '-' && next+1 < len(spec) {
			end, rangeNext, rangeSpecial, ok := parseASCIICharClassAtom(spec, next+1)
			if !ok || rangeSpecial != 0 || end < ch {
				return nil, false
			}
			for b := ch; b <= end; b++ {
				class.Table[b] = true
			}
			i = rangeNext - 1
			continue
		}
		class.Table[ch] = true
		i = next - 1
	}
	if negated {
		for i := range class.Table {
			class.Table[i] = !class.Table[i]
		}
	}
	return class, true
}

func parseASCIICharClassAtom(spec string, pos int) (byte, int, byte, bool) {
	if pos >= len(spec) || spec[pos] >= utf8.RuneSelf {
		return 0, 0, 0, false
	}
	if spec[pos] != '\\' {
		return spec[pos], pos + 1, 0, true
	}
	if pos+1 >= len(spec) || spec[pos+1] >= utf8.RuneSelf {
		return 0, 0, 0, false
	}
	esc := spec[pos+1]
	switch esc {
	case 'w', 'W', 'd', 'D', 's', 'S':
		return 0, pos + 2, esc, true
	case 't':
		return '\t', pos + 2, 0, true
	case 'n':
		return '\n', pos + 2, 0, true
	case 'r':
		return '\r', pos + 2, 0, true
	case 'f':
		return '\f', pos + 2, 0, true
	case 'v':
		return '\v', pos + 2, 0, true
	case 'a':
		return '\a', pos + 2, 0, true
	case 'x', 'u', 'U', 'p', 'P':
		return 0, 0, 0, false
	default:
		if (esc >= 'A' && esc <= 'Z') || (esc >= 'a' && esc <= 'z') {
			return 0, 0, 0, false
		}
		return esc, pos + 2, 0, true
	}
}

func addASCIICharClassSpecial(class *asciiCharClass, special byte) {
	switch special {
	case 'w':
		for b := 0; b < 256; b++ {
			if isWordByte(byte(b)) {
				class.Table[b] = true
			}
		}
	case 'W':
		for b := 0; b < 256; b++ {
			if !isWordByte(byte(b)) {
				class.Table[b] = true
			}
		}
	case 'd':
		for ch := byte('0'); ch <= '9'; ch++ {
			class.Table[ch] = true
		}
	case 'D':
		for b := 0; b < 256; b++ {
			if b < '0' || b > '9' {
				class.Table[b] = true
			}
		}
	case 's':
		for b := 0; b < 256; b++ {
			if isRegexpASCIISpace(byte(b)) {
				class.Table[b] = true
			}
		}
	case 'S':
		for b := 0; b < 256; b++ {
			if !isRegexpASCIISpace(byte(b)) {
				class.Table[b] = true
			}
		}
	}
}

func nextStructuredSlicingLiteral(steps []structuredStep, start int) string {
	for i := start; i < len(steps); i++ {
		step := steps[i]
		if step.optional {
			return ""
		}
		if step.literal != "" {
			return step.literal
		}
		if step.parser != nil {
			if step.parser.wrapPrefix != "" {
				return step.parser.wrapPrefix
			}
			switch step.parser.kind {
			case structuredSpaceOne, structuredSpacePlus, structuredSpaceStar:
				continue
			default:
				return ""
			}
		}
		return ""
	}
	return ""
}

func nextStructuredParserKind(steps []structuredStep, start int) (structuredKind, bool) {
	for i := start; i < len(steps); i++ {
		if steps[i].literal != "" {
			return 0, false
		}
		if steps[i].parser != nil {
			return steps[i].parser.kind, true
		}
		return 0, false
	}
	return 0, false
}

func firstStructuredLiteral(step structuredStep) (string, bool) {
	switch {
	case step.literal != "":
		return step.literal, true
	case step.submatcher != nil:
		return step.submatcher.ir.FirstLiteral, step.submatcher.ir.FirstLiteralExact && step.submatcher.ir.FirstLiteral != ""
	case len(step.alternatives) > 0:
		return step.ir.FirstLiteral, step.ir.FirstLiteralExact && step.ir.FirstLiteral != ""
	case step.parser != nil && step.parser.wrapPrefix != "":
		return step.parser.wrapPrefix, true
	default:
		return "", false
	}
}

func (m structuredMatcher) match(dst []string, content string, trimSpace bool) bool {
	if m.anchoredStart || m.prefersStartOnly() {
		return m.matchTopAt(dst, content, 0, trimSpace)
	}
	if m.riskySearch {
		resetStringResults(dst)
		return m.matchTopAt(dst, content, 0, trimSpace)
	}
	if m.hasStartOnlyRunner() {
		resetStringResults(dst)
		if m.matchTopAt(dst, content, 0, trimSpace) {
			return true
		}
	}
	return m.matchSearch(dst, content, trimSpace)
}

func (m structuredMatcher) matchTyped(dst []any, content string, trimSpace bool, kinds []valueKind) bool {
	if m.anchoredStart || m.prefersStartOnly() {
		return m.matchTypedTopAt(dst, content, 0, trimSpace, kinds)
	}
	if m.riskySearch {
		resetAnyResults(dst)
		return m.matchTypedTopAt(dst, content, 0, trimSpace, kinds)
	}
	if m.hasStartOnlyRunner() {
		resetAnyResults(dst)
		if m.matchTypedTopAt(dst, content, 0, trimSpace, kinds) {
			return true
		}
	}
	return m.matchTypedSearch(dst, content, trimSpace, kinds)
}

func (m structuredMatcher) hasStartOnlyRunner() bool {
	return m.startOnlyRunner
}

func (m structuredMatcher) hasConcreteStartOnlyRunner() bool {
	return m.accessRunner != nil ||
		m.commonRunner != nil ||
		m.jenkinsRunner != nil ||
		m.tomcatRunner != nil ||
		m.mysqlRunner != nil ||
		m.sqlserverRunner != nil ||
		m.kafkaRunner != nil ||
		m.kingbaseRunner != nil ||
		m.redisRunner != nil ||
		m.damengRunner != nil ||
		m.elasticRunner != nil ||
		m.elasticDefaultRunner != nil ||
		m.elasticSearchRunner != nil ||
		m.nginxErrorRunner != nil ||
		m.postfixRunner != nil ||
		m.rabbitRunner != nil ||
		m.solrRunner != nil
}

func (m structuredMatcher) prefersStartOnly() bool {
	if len(m.steps) == 0 {
		return true
	}
	step := m.steps[0]
	if step.optional {
		return false
	}
	if step.parser == nil {
		return false
	}
	switch step.parser.kind {
	case structuredUntilLiteral, structuredGreedyUntilLiteral, structuredSpaceStar:
		return true
	default:
		return false
	}
}

func (m structuredMatcher) matchTopAt(dst []string, content string, pos int, trimSpace bool) bool {
	if m.accessRunner != nil {
		if m.accessRunner.run(dst, content[pos:], trimSpace) {
			return true
		}
		resetStringResults(dst)
		return false
	}
	if m.tomcatRunner != nil {
		if m.tomcatRunner.run(dst, content[pos:], trimSpace) {
			return true
		}
		resetStringResults(dst)
		return false
	}
	if m.postfixRunner != nil {
		if m.postfixRunner.run(dst, content[pos:], trimSpace) {
			return true
		}
		resetStringResults(dst)
		return false
	}
	if pos == 0 && m.startOnlyRunner {
		if pos == 0 && m.commonRunner != nil {
			if m.commonRunner.run(dst, content, trimSpace) {
				return true
			}
			resetStringResults(dst)
		}
		if pos == 0 && m.jenkinsRunner != nil {
			if m.jenkinsRunner.run(dst, content, trimSpace) {
				return true
			}
			resetStringResults(dst)
		}
		if pos == 0 && m.mysqlRunner != nil {
			if m.mysqlRunner.run(dst, content, trimSpace) {
				return true
			}
			resetStringResults(dst)
		}
		if pos == 0 && m.sqlserverRunner != nil {
			if m.sqlserverRunner.run(dst, content, trimSpace) {
				return true
			}
			resetStringResults(dst)
		}
		if pos == 0 && m.kafkaRunner != nil {
			if m.kafkaRunner.run(dst, content, trimSpace) {
				return true
			}
			resetStringResults(dst)
		}
		if pos == 0 && m.kingbaseRunner != nil {
			if m.kingbaseRunner.run(dst, content, trimSpace) {
				return true
			}
			resetStringResults(dst)
		}
		if pos == 0 && m.redisRunner != nil {
			if m.redisRunner.run(dst, content, trimSpace) {
				return true
			}
			resetStringResults(dst)
		}
		if pos == 0 && m.damengRunner != nil {
			if m.damengRunner.run(dst, content, trimSpace) {
				return true
			}
			resetStringResults(dst)
		}
		if pos == 0 && m.elasticRunner != nil {
			if m.elasticRunner.run(dst, content, trimSpace) {
				return true
			}
			resetStringResults(dst)
		}
		if pos == 0 && m.elasticDefaultRunner != nil {
			if m.elasticDefaultRunner.run(dst, content, trimSpace) {
				return true
			}
			resetStringResults(dst)
		}
		if pos == 0 && m.elasticSearchRunner != nil {
			if m.elasticSearchRunner.run(dst, content, trimSpace) {
				return true
			}
			resetStringResults(dst)
		}
		if pos == 0 && m.nginxErrorRunner != nil {
			if m.nginxErrorRunner.run(dst, content, trimSpace) {
				return true
			}
			resetStringResults(dst)
		}
		if pos == 0 && m.rabbitRunner != nil {
			if m.rabbitRunner.run(dst, content, trimSpace) {
				return true
			}
			resetStringResults(dst)
		}
		if pos == 0 && m.solrRunner != nil {
			if m.solrRunner.run(dst, content, trimSpace) {
				return true
			}
			resetStringResults(dst)
		}
	}
	if m.quickReject(content, pos) {
		return false
	}
	if m.anchoredRunner != nil {
		next, ok := m.anchoredRunner.runAt(dst, content, pos, trimSpace)
		return ok && m.matchEndOK(next, content)
	}
	if !m.backtracking && !m.changeLog {
		next, ok := m.matchLinearFrom(dst, content, pos, trimSpace)
		return ok && m.matchEndOK(next, content)
	}

	changeCap := len(dst)
	if m.changeCap > changeCap {
		changeCap = m.changeCap
	}
	changeBuf := getMatchChangeBuffer(changeCap)
	changes := changeBuf.changes

	var next int
	var ok bool
	if m.backtracking {
		next, ok, changes = m.matchBacktrackingFrom(dst, content, pos, trimSpace, changes)
	} else {
		next, ok, changes = m.matchFrom(dst, content, pos, trimSpace, changes)
	}
	putMatchChangeBuffer(changeBuf, changes)
	return ok && m.matchEndOK(next, content)
}

func (m structuredMatcher) matchTypedTopAt(dst []any, content string, pos int, trimSpace bool, kinds []valueKind) bool {
	if m.accessRunner != nil {
		if m.accessRunner.runTyped(dst, content[pos:], trimSpace, kinds) {
			return true
		}
		resetAnyResults(dst)
		return false
	}
	if m.tomcatRunner != nil {
		if m.tomcatRunner.runTyped(dst, content[pos:], trimSpace, kinds) {
			return true
		}
		resetAnyResults(dst)
		return false
	}
	if m.postfixRunner != nil {
		if m.postfixRunner.runTyped(dst, content[pos:], trimSpace, kinds) {
			return true
		}
		resetAnyResults(dst)
		return false
	}
	if pos == 0 && m.startOnlyRunner {
		if pos == 0 && m.commonRunner != nil {
			if m.commonRunner.runTyped(dst, content, trimSpace, kinds) {
				return true
			}
			resetAnyResults(dst)
		}
		if pos == 0 && m.jenkinsRunner != nil {
			if m.jenkinsRunner.runTyped(dst, content, trimSpace, kinds) {
				return true
			}
			resetAnyResults(dst)
		}
		if pos == 0 && m.mysqlRunner != nil {
			if m.mysqlRunner.runTyped(dst, content, trimSpace, kinds) {
				return true
			}
			resetAnyResults(dst)
		}
		if pos == 0 && m.sqlserverRunner != nil {
			if m.sqlserverRunner.runTyped(dst, content, trimSpace, kinds) {
				return true
			}
			resetAnyResults(dst)
		}
		if pos == 0 && m.kafkaRunner != nil {
			if m.kafkaRunner.runTyped(dst, content, trimSpace, kinds) {
				return true
			}
			resetAnyResults(dst)
		}
		if pos == 0 && m.kingbaseRunner != nil {
			if m.kingbaseRunner.runTyped(dst, content, trimSpace, kinds) {
				return true
			}
			resetAnyResults(dst)
		}
		if pos == 0 && m.redisRunner != nil {
			if m.redisRunner.runTyped(dst, content, trimSpace, kinds) {
				return true
			}
			resetAnyResults(dst)
		}
		if pos == 0 && m.damengRunner != nil {
			if m.damengRunner.runTyped(dst, content, trimSpace, kinds) {
				return true
			}
			resetAnyResults(dst)
		}
		if pos == 0 && m.elasticRunner != nil {
			if m.elasticRunner.runTyped(dst, content, trimSpace, kinds) {
				return true
			}
			resetAnyResults(dst)
		}
		if pos == 0 && m.elasticDefaultRunner != nil {
			if m.elasticDefaultRunner.runTyped(dst, content, trimSpace, kinds) {
				return true
			}
			resetAnyResults(dst)
		}
		if pos == 0 && m.elasticSearchRunner != nil {
			if m.elasticSearchRunner.runTyped(dst, content, trimSpace, kinds) {
				return true
			}
			resetAnyResults(dst)
		}
		if pos == 0 && m.nginxErrorRunner != nil {
			if m.nginxErrorRunner.runTyped(dst, content, trimSpace, kinds) {
				return true
			}
			resetAnyResults(dst)
		}
		if pos == 0 && m.rabbitRunner != nil {
			if m.rabbitRunner.runTyped(dst, content, trimSpace, kinds) {
				return true
			}
			resetAnyResults(dst)
		}
		if pos == 0 && m.solrRunner != nil {
			if m.solrRunner.runTyped(dst, content, trimSpace, kinds) {
				return true
			}
			resetAnyResults(dst)
		}
	}
	resetTypedResults(dst, kinds)
	if m.quickReject(content, pos) {
		return false
	}
	if m.anchoredRunner != nil {
		next, ok := m.anchoredRunner.runTypedAt(dst, content, pos, trimSpace, kinds)
		return ok && m.matchEndOK(next, content)
	}
	if !m.backtracking && !m.changeLog {
		next, ok := m.matchTypedLinearFrom(dst, content, pos, trimSpace, kinds)
		return ok && m.matchEndOK(next, content)
	}

	changeCap := len(dst)
	if m.changeCap > changeCap {
		changeCap = m.changeCap
	}
	changeBuf := getTypedMatchChangeBuffer(changeCap)
	changes := changeBuf.changes

	var next int
	var ok bool
	if m.backtracking {
		next, ok, changes = m.matchTypedBacktrackingFrom(dst, content, pos, trimSpace, kinds, changes)
	} else {
		next, ok, changes = m.matchTypedFrom(dst, content, pos, trimSpace, kinds, changes)
	}
	putTypedMatchChangeBuffer(changeBuf, changes)
	return ok && m.matchEndOK(next, content)
}

func (m structuredMatcher) matchEndOK(next int, content string) bool {
	if next < 0 {
		return false
	}
	if !m.anchoredEnd {
		return true
	}
	return next == len(content)
}

func (m structuredMatcher) matchSearch(dst []string, content string, trimSpace bool) bool {
	for pos := m.nextSearchPos(content, 0); pos <= len(content); pos = m.nextSearchPos(content, pos+1) {
		resetStringResults(dst)
		if m.matchTopAt(dst, content, pos, trimSpace) {
			return true
		}
		if len(content)-pos < m.ir.MinWidth {
			break
		}
	}
	return false
}

func (m structuredMatcher) matchTypedSearch(dst []any, content string, trimSpace bool, kinds []valueKind) bool {
	for pos := m.nextSearchPos(content, 0); pos <= len(content); pos = m.nextSearchPos(content, pos+1) {
		resetAnyResults(dst)
		if m.matchTypedTopAt(dst, content, pos, trimSpace, kinds) {
			return true
		}
		if len(content)-pos < m.ir.MinWidth {
			break
		}
	}
	return false
}

func (m structuredMatcher) nextSearchPos(content string, pos int) int {
	if pos <= 0 {
		if m.ir.FirstLiteralExact && m.ir.FirstLiteral != "" {
			idx := strings.Index(content, m.ir.FirstLiteral)
			if idx < 0 {
				return len(content) + 1
			}
			return idx
		}
		return 0
	}
	if pos > len(content) {
		return len(content) + 1
	}
	if m.ir.FirstLiteralExact && m.ir.FirstLiteral != "" {
		idx := strings.Index(content[pos:], m.ir.FirstLiteral)
		if idx < 0 {
			return len(content) + 1
		}
		return pos + idx
	}
	if pos == len(content) {
		return pos
	}
	if content[pos] < utf8.RuneSelf {
		return pos
	}
	_, size := utf8.DecodeRuneInString(content[pos:])
	if size <= 0 {
		return pos
	}
	return pos
}

func resetStringResults(dst []string) {
	for i := range dst {
		dst[i] = ""
	}
}

func resetAnyResults(dst []any) {
	for i := range dst {
		dst[i] = nil
	}
}

func resetTypedResults(dst []any, kinds []valueKind) {
	for i := range dst {
		dst[i], _ = castValue("", kinds[i])
	}
}

func (m structuredMatcher) matchAnyFrom(dst []string, content string, pos int, trimSpace bool, changes []matchChange) (int, bool, []matchChange) {
	if m.quickReject(content, pos) {
		return 0, false, changes
	}
	if !m.backtracking && !m.changeLog && changes == nil {
		next, ok := m.matchLinearFrom(dst, content, pos, trimSpace)
		return next, ok, nil
	}
	if m.backtracking {
		return m.matchBacktrackingFrom(dst, content, pos, trimSpace, changes)
	}
	return m.matchFrom(dst, content, pos, trimSpace, changes)
}

func (m structuredMatcher) matchLinearFrom(dst []string, content string, pos int, trimSpace bool) (int, bool) {
	if m.quickReject(content, pos) {
		return 0, false
	}
	for _, step := range m.steps {
		var next int
		var ok bool

		switch {
		case step.literal != "":
			if !strings.HasPrefix(content[pos:], step.literal) {
				if step.optional {
					if step.deterministicOptional && !step.optPrefixSkips && step.optPrefix != "" && strings.HasPrefix(content[pos:], step.optPrefix) {
						return 0, false
					}
					continue
				}
				return 0, false
			}
			if step.captureIndex >= 0 {
				dst[step.captureIndex] = maybeTrim(content[pos:pos+len(step.literal)], trimSpace)
			}
			pos += len(step.literal)
			continue
		case step.parser != nil:
			if step.optional && step.deterministicOptional && shouldSkipOptionalParser(step.parser, step, content, pos) {
				continue
			}
			var value string
			next, value, ok = step.parser.consume(content, pos)
			if !ok {
				if step.optional {
					if step.deterministicOptional {
						return 0, false
					}
					continue
				}
				return 0, false
			}
			if step.parser.dstIndex >= 0 {
				dst[step.parser.dstIndex] = maybeTrim(value, trimSpace)
			}
		case step.submatcher != nil:
			next, ok = step.submatcher.matchLinearFrom(dst, content, pos, trimSpace)
			if !ok {
				if step.optional {
					if step.deterministicOptional && !step.optPrefixSkips && step.optPrefix != "" && strings.HasPrefix(content[pos:], step.optPrefix) {
						return 0, false
					}
					continue
				}
				return 0, false
			}
			if step.captureIndex >= 0 {
				dst[step.captureIndex] = maybeTrim(content[pos:next], trimSpace)
			}
		case len(step.alternatives) > 0:
			alts := matchingAlternatives(step, content, pos)
			for _, alt := range alts {
				next, ok = alt.matchLinearFrom(dst, content, pos, trimSpace)
				if ok {
					if step.captureIndex >= 0 {
						dst[step.captureIndex] = maybeTrim(content[pos:next], trimSpace)
					}
					break
				}
			}
			if !ok {
				if step.optional {
					if step.deterministicOptional && !step.optPrefixSkips && step.optPrefix != "" && strings.HasPrefix(content[pos:], step.optPrefix) {
						return 0, false
					}
					continue
				}
				return 0, false
			}
		default:
			continue
		}

		pos = next
	}

	return pos, true
}

func (m structuredMatcher) matchTypedLinearFrom(dst []any, content string, pos int, trimSpace bool, kinds []valueKind) (int, bool) {
	if m.quickReject(content, pos) {
		return 0, false
	}
	for _, step := range m.steps {
		var next int
		var ok bool

		switch {
		case step.literal != "":
			if !strings.HasPrefix(content[pos:], step.literal) {
				if step.optional {
					if step.deterministicOptional && !step.optPrefixSkips && step.optPrefix != "" && strings.HasPrefix(content[pos:], step.optPrefix) {
						return 0, false
					}
					continue
				}
				return 0, false
			}
			if step.captureIndex >= 0 {
				dst[step.captureIndex] = castStructuredValue(content[pos:pos+len(step.literal)], trimSpace, kinds[step.captureIndex])
			}
			pos += len(step.literal)
			continue
		case step.parser != nil:
			if step.optional && step.deterministicOptional && shouldSkipOptionalParser(step.parser, step, content, pos) {
				continue
			}
			var value string
			next, value, ok = step.parser.consume(content, pos)
			if !ok {
				if step.optional {
					if step.deterministicOptional {
						return 0, false
					}
					continue
				}
				return 0, false
			}
			if step.parser.dstIndex >= 0 {
				dst[step.parser.dstIndex] = castStructuredValue(value, trimSpace, kinds[step.parser.dstIndex])
			}
		case step.submatcher != nil:
			next, ok = step.submatcher.matchTypedLinearFrom(dst, content, pos, trimSpace, kinds)
			if !ok {
				if step.optional {
					if step.deterministicOptional && !step.optPrefixSkips && step.optPrefix != "" && strings.HasPrefix(content[pos:], step.optPrefix) {
						return 0, false
					}
					continue
				}
				return 0, false
			}
			if step.captureIndex >= 0 {
				dst[step.captureIndex] = castStructuredValue(content[pos:next], trimSpace, kinds[step.captureIndex])
			}
		case len(step.alternatives) > 0:
			alts := matchingAlternatives(step, content, pos)
			for _, alt := range alts {
				next, ok = alt.matchTypedLinearFrom(dst, content, pos, trimSpace, kinds)
				if ok {
					if step.captureIndex >= 0 {
						dst[step.captureIndex] = castStructuredValue(content[pos:next], trimSpace, kinds[step.captureIndex])
					}
					break
				}
			}
			if !ok {
				if step.optional {
					if step.deterministicOptional && !step.optPrefixSkips && step.optPrefix != "" && strings.HasPrefix(content[pos:], step.optPrefix) {
						return 0, false
					}
					continue
				}
				return 0, false
			}
		default:
			continue
		}

		pos = next
	}

	return pos, true
}

func shouldSkipOptionalParser(p *structuredParser, step structuredStep, content string, pos int) bool {
	if p == nil || pos > len(content) {
		return false
	}
	if step.optPrefixSkips && step.optPrefix != "" && strings.HasPrefix(content[pos:], step.optPrefix) {
		return true
	}
	if p.dstIndex < 0 && p.kind == structuredYear && p.hasNextParser && p.nextParser == structuredTimeOfDay {
		_, _, ok := sliceTimeOfDay(content, pos)
		return ok
	}
	return false
}

func (m structuredMatcher) matchFrom(dst []string, content string, pos int, trimSpace bool, changes []matchChange) (int, bool, []matchChange) {
	if m.quickReject(content, pos) {
		return 0, false, changes
	}
	for _, step := range m.steps {
		var next int
		var ok bool

		switch {
		case step.literal != "":
			if !strings.HasPrefix(content[pos:], step.literal) {
				if step.optional {
					continue
				}
				return 0, false, changes
			}
			if step.captureIndex >= 0 {
				changes = appendMatchChange(dst, changes, step.captureIndex, maybeTrim(content[pos:pos+len(step.literal)], trimSpace))
			}
			pos += len(step.literal)
			continue
		case step.parser != nil:
			var value string
			next, value, ok = step.parser.consume(content, pos)
			if !ok {
				if step.optional {
					continue
				}
				return 0, false, changes
			}
			if step.parser.dstIndex >= 0 {
				changes = appendMatchChange(dst, changes, step.parser.dstIndex, maybeTrim(value, trimSpace))
			}
		case step.submatcher != nil:
			mark := len(changes)
			if step.submatcher.backtracking {
				next, ok, changes = step.submatcher.matchBacktrackingFrom(dst, content, pos, trimSpace, changes)
			} else {
				next, ok, changes = step.submatcher.matchFrom(dst, content, pos, trimSpace, changes)
			}
			if !ok {
				if step.submatcher.writes {
					rollbackMatchChanges(dst, changes, mark)
					changes = changes[:mark]
				}
				if step.optional {
					continue
				}
				return 0, false, changes
			}
			if step.captureIndex >= 0 {
				changes = appendMatchChange(dst, changes, step.captureIndex, maybeTrim(content[pos:next], trimSpace))
			}
		case len(step.alternatives) > 0:
			alts := matchingAlternatives(step, content, pos)
			for _, alt := range alts {
				mark := len(changes)
				if alt.backtracking {
					next, ok, changes = alt.matchBacktrackingFrom(dst, content, pos, trimSpace, changes)
				} else {
					next, ok, changes = alt.matchFrom(dst, content, pos, trimSpace, changes)
				}
				if ok {
					if step.captureIndex >= 0 {
						changes = appendMatchChange(dst, changes, step.captureIndex, maybeTrim(content[pos:next], trimSpace))
					}
					break
				}
				if alt.writes {
					rollbackMatchChanges(dst, changes, mark)
					changes = changes[:mark]
				}
			}
			if !ok {
				if step.optional {
					continue
				}
				return 0, false, changes
			}
		default:
			continue
		}

		pos = next
	}

	return pos, true, changes
}

func (m structuredMatcher) matchTypedFrom(dst []any, content string, pos int, trimSpace bool, kinds []valueKind, changes []typedMatchChange) (int, bool, []typedMatchChange) {
	if m.quickReject(content, pos) {
		return 0, false, changes
	}
	for _, step := range m.steps {
		var next int
		var ok bool

		switch {
		case step.literal != "":
			if !strings.HasPrefix(content[pos:], step.literal) {
				if step.optional {
					continue
				}
				return 0, false, changes
			}
			if step.captureIndex >= 0 {
				changes = appendTypedMatchChange(dst, changes, step.captureIndex, castStructuredValue(content[pos:pos+len(step.literal)], trimSpace, kinds[step.captureIndex]))
			}
			pos += len(step.literal)
			continue
		case step.parser != nil:
			var value string
			next, value, ok = step.parser.consume(content, pos)
			if !ok {
				if step.optional {
					continue
				}
				return 0, false, changes
			}
			if step.parser.dstIndex >= 0 {
				changes = appendTypedMatchChange(dst, changes, step.parser.dstIndex, castStructuredValue(value, trimSpace, kinds[step.parser.dstIndex]))
			}
		case step.submatcher != nil:
			mark := len(changes)
			if step.submatcher.backtracking {
				next, ok, changes = step.submatcher.matchTypedBacktrackingFrom(dst, content, pos, trimSpace, kinds, changes)
			} else {
				next, ok, changes = step.submatcher.matchTypedFrom(dst, content, pos, trimSpace, kinds, changes)
			}
			if !ok {
				if step.submatcher.writes {
					rollbackTypedMatchChanges(dst, changes, mark)
					changes = changes[:mark]
				}
				if step.optional {
					continue
				}
				return 0, false, changes
			}
			if step.captureIndex >= 0 {
				changes = appendTypedMatchChange(dst, changes, step.captureIndex, castStructuredValue(content[pos:next], trimSpace, kinds[step.captureIndex]))
			}
		case len(step.alternatives) > 0:
			alts := matchingAlternatives(step, content, pos)
			for _, alt := range alts {
				mark := len(changes)
				if alt.backtracking {
					next, ok, changes = alt.matchTypedBacktrackingFrom(dst, content, pos, trimSpace, kinds, changes)
				} else {
					next, ok, changes = alt.matchTypedFrom(dst, content, pos, trimSpace, kinds, changes)
				}
				if ok {
					if step.captureIndex >= 0 {
						changes = appendTypedMatchChange(dst, changes, step.captureIndex, castStructuredValue(content[pos:next], trimSpace, kinds[step.captureIndex]))
					}
					break
				}
				if alt.writes {
					rollbackTypedMatchChanges(dst, changes, mark)
					changes = changes[:mark]
				}
			}
			if !ok {
				if step.optional {
					continue
				}
				return 0, false, changes
			}
		default:
			continue
		}

		pos = next
	}

	return pos, true, changes
}

func (m structuredMatcher) matchBacktrackingFrom(dst []string, content string, pos int, trimSpace bool, changes []matchChange) (int, bool, []matchChange) {
	return matchStructuredSteps(m.steps, 0, dst, content, pos, trimSpace, changes)
}

func (m structuredMatcher) matchTypedBacktrackingFrom(dst []any, content string, pos int, trimSpace bool, kinds []valueKind, changes []typedMatchChange) (int, bool, []typedMatchChange) {
	return matchTypedStructuredSteps(m.steps, 0, dst, content, pos, trimSpace, kinds, changes)
}

func matchStructuredSteps(steps []structuredStep, idx int, dst []string, content string, pos int, trimSpace bool, changes []matchChange) (int, bool, []matchChange) {
	if idx >= len(steps) {
		return pos, true, changes
	}
	if len(content)-pos < steps[idx].tailMinWidth {
		return 0, false, changes
	}

	step := steps[idx]
	if step.parserBacktracking {
		return matchStructuredBacktrackingParserStep(steps, idx, step, dst, content, pos, trimSpace, changes)
	}

	mark := len(changes)
	if step.optional {
		return matchStructuredOptionalStep(steps, idx, step, dst, content, pos, trimSpace, changes, mark)
	}

	next, ok, nextChanges := matchStructuredStep(step, dst, content, pos, trimSpace, changes)
	if !ok {
		return 0, false, changes[:mark]
	}

	if end, okEnd, endChanges := matchStructuredSteps(steps, idx+1, dst, content, next, trimSpace, nextChanges); okEnd {
		return end, true, endChanges
	}

	if step.writes {
		rollbackMatchChanges(dst, nextChanges, mark)
	}
	return 0, false, changes[:mark]
}

func matchTypedStructuredSteps(steps []structuredStep, idx int, dst []any, content string, pos int, trimSpace bool, kinds []valueKind, changes []typedMatchChange) (int, bool, []typedMatchChange) {
	if idx >= len(steps) {
		return pos, true, changes
	}
	if len(content)-pos < steps[idx].tailMinWidth {
		return 0, false, changes
	}

	step := steps[idx]
	if step.parserBacktracking {
		return matchTypedBacktrackingParserStep(steps, idx, step, dst, content, pos, trimSpace, kinds, changes)
	}

	mark := len(changes)
	if step.optional {
		return matchTypedOptionalStep(steps, idx, step, dst, content, pos, trimSpace, kinds, changes, mark)
	}

	next, ok, nextChanges := matchTypedStructuredStep(step, dst, content, pos, trimSpace, kinds, changes)
	if !ok {
		return 0, false, changes[:mark]
	}

	if end, okEnd, endChanges := matchTypedStructuredSteps(steps, idx+1, dst, content, next, trimSpace, kinds, nextChanges); okEnd {
		return end, true, endChanges
	}

	if step.writes {
		rollbackTypedMatchChanges(dst, nextChanges, mark)
	}
	return 0, false, changes[:mark]
}

func matchStructuredBacktrackingParserStep(steps []structuredStep, idx int, step structuredStep, dst []string, content string, pos int, trimSpace bool, changes []matchChange) (int, bool, []matchChange) {
	mark := len(changes)
	rest := content[pos:]
	end := len(rest)

	for {
		rel := lastLiteralIndex(rest[:end], step.parser.nextLiteral)
		if rel < 0 {
			break
		}

		nextChanges := changes
		segment := rest[:rel]
		if !backtrackingParserSegmentOK(step.parser, segment) {
			end = rel
			continue
		}
		if step.parser.dstIndex >= 0 {
			nextChanges = appendMatchChange(dst, nextChanges, step.parser.dstIndex, maybeTrim(segment, trimSpace))
		}

		if matchEnd, ok, endChanges := matchStructuredSteps(steps, idx+1, dst, content, pos+rel, trimSpace, nextChanges); ok {
			return matchEnd, true, endChanges
		}
		if len(nextChanges) > mark {
			rollbackMatchChanges(dst, nextChanges, mark)
		}

		end = rel
	}

	if step.optional {
		return matchStructuredSteps(steps, idx+1, dst, content, pos, trimSpace, changes[:mark])
	}

	return 0, false, changes[:mark]
}

func matchTypedBacktrackingParserStep(steps []structuredStep, idx int, step structuredStep, dst []any, content string, pos int, trimSpace bool, kinds []valueKind, changes []typedMatchChange) (int, bool, []typedMatchChange) {
	mark := len(changes)
	rest := content[pos:]
	end := len(rest)

	for {
		rel := lastLiteralIndex(rest[:end], step.parser.nextLiteral)
		if rel < 0 {
			break
		}

		nextChanges := changes
		segment := rest[:rel]
		if !backtrackingParserSegmentOK(step.parser, segment) {
			end = rel
			continue
		}
		if step.parser.dstIndex >= 0 {
			nextChanges = appendTypedMatchChange(dst, nextChanges, step.parser.dstIndex, castStructuredValue(segment, trimSpace, kinds[step.parser.dstIndex]))
		}

		if matchEnd, ok, endChanges := matchTypedStructuredSteps(steps, idx+1, dst, content, pos+rel, trimSpace, kinds, nextChanges); ok {
			return matchEnd, true, endChanges
		}
		if len(nextChanges) > mark {
			rollbackTypedMatchChanges(dst, nextChanges, mark)
		}

		end = rel
	}

	if step.optional {
		return matchTypedStructuredSteps(steps, idx+1, dst, content, pos, trimSpace, kinds, changes[:mark])
	}

	return 0, false, changes[:mark]
}

func backtrackingParserSegmentOK(p *structuredParser, segment string) bool {
	if p == nil {
		return false
	}
	switch p.kind {
	case structuredGreedyUntilLiteral:
		return dotRunSegmentOK(p, segment)
	case structuredNotSpace:
		return segment != "" && strings.IndexAny(segment, " \t\r\n\f\v") < 0
	default:
		return false
	}
}

func matchStructuredOptionalStep(steps []structuredStep, idx int, step structuredStep, dst []string, content string, pos int, trimSpace bool, changes []matchChange, mark int) (int, bool, []matchChange) {
	next, ok, nextChanges := matchStructuredStep(step, dst, content, pos, trimSpace, changes)
	if ok {
		if end, okEnd, endChanges := matchStructuredSteps(steps, idx+1, dst, content, next, trimSpace, nextChanges); okEnd {
			return end, true, endChanges
		}
		if len(nextChanges) > mark {
			rollbackMatchChanges(dst, nextChanges, mark)
		}
	}

	return matchStructuredSteps(steps, idx+1, dst, content, pos, trimSpace, changes[:mark])
}

func matchTypedOptionalStep(steps []structuredStep, idx int, step structuredStep, dst []any, content string, pos int, trimSpace bool, kinds []valueKind, changes []typedMatchChange, mark int) (int, bool, []typedMatchChange) {
	next, ok, nextChanges := matchTypedStructuredStep(step, dst, content, pos, trimSpace, kinds, changes)
	if ok {
		if end, okEnd, endChanges := matchTypedStructuredSteps(steps, idx+1, dst, content, next, trimSpace, kinds, nextChanges); okEnd {
			return end, true, endChanges
		}
		if len(nextChanges) > mark {
			rollbackTypedMatchChanges(dst, nextChanges, mark)
		}
	}

	return matchTypedStructuredSteps(steps, idx+1, dst, content, pos, trimSpace, kinds, changes[:mark])
}

func matchStructuredStep(step structuredStep, dst []string, content string, pos int, trimSpace bool, changes []matchChange) (int, bool, []matchChange) {
	switch {
	case step.literal != "":
		if !strings.HasPrefix(content[pos:], step.literal) {
			return 0, false, changes
		}
		if step.captureIndex >= 0 {
			changes = appendMatchChange(dst, changes, step.captureIndex, maybeTrim(content[pos:pos+len(step.literal)], trimSpace))
		}
		return pos + len(step.literal), true, changes
	case step.parser != nil:
		next, value, ok := step.parser.consume(content, pos)
		if !ok {
			return 0, false, changes
		}
		if step.parser.dstIndex >= 0 {
			changes = appendMatchChange(dst, changes, step.parser.dstIndex, maybeTrim(value, trimSpace))
		}
		return next, true, changes
	case step.submatcher != nil:
		mark := len(changes)
		next, ok, nextChanges := step.submatcher.matchAnyFrom(dst, content, pos, trimSpace, changes)
		if !ok {
			if step.submatcher.writes {
				rollbackMatchChanges(dst, nextChanges, mark)
				nextChanges = nextChanges[:mark]
			}
			return 0, false, changes
		}
		if step.captureIndex >= 0 {
			nextChanges = appendMatchChange(dst, nextChanges, step.captureIndex, maybeTrim(content[pos:next], trimSpace))
		}
		return next, true, nextChanges
	case len(step.alternatives) > 0:
		for _, alt := range matchingAlternatives(step, content, pos) {
			mark := len(changes)
			next, ok, nextChanges := alt.matchAnyFrom(dst, content, pos, trimSpace, changes)
			if ok {
				if step.captureIndex >= 0 {
					nextChanges = appendMatchChange(dst, nextChanges, step.captureIndex, maybeTrim(content[pos:next], trimSpace))
				}
				return next, true, nextChanges
			}
			if alt.writes {
				rollbackMatchChanges(dst, nextChanges, mark)
				nextChanges = nextChanges[:mark]
			}
		}
		return 0, false, changes
	default:
		return pos, true, changes
	}
}

func matchTypedStructuredStep(step structuredStep, dst []any, content string, pos int, trimSpace bool, kinds []valueKind, changes []typedMatchChange) (int, bool, []typedMatchChange) {
	switch {
	case step.literal != "":
		if !strings.HasPrefix(content[pos:], step.literal) {
			return 0, false, changes
		}
		if step.captureIndex >= 0 {
			changes = appendTypedMatchChange(dst, changes, step.captureIndex, castStructuredValue(content[pos:pos+len(step.literal)], trimSpace, kinds[step.captureIndex]))
		}
		return pos + len(step.literal), true, changes
	case step.parser != nil:
		next, value, ok := step.parser.consume(content, pos)
		if !ok {
			return 0, false, changes
		}
		if step.parser.dstIndex >= 0 {
			changes = appendTypedMatchChange(dst, changes, step.parser.dstIndex, castStructuredValue(value, trimSpace, kinds[step.parser.dstIndex]))
		}
		return next, true, changes
	case step.submatcher != nil:
		mark := len(changes)
		next, ok, nextChanges := step.submatcher.matchTypedAnyFrom(dst, content, pos, trimSpace, kinds, changes)
		if !ok {
			if step.submatcher.writes {
				rollbackTypedMatchChanges(dst, nextChanges, mark)
				nextChanges = nextChanges[:mark]
			}
			return 0, false, changes
		}
		if step.captureIndex >= 0 {
			nextChanges = appendTypedMatchChange(dst, nextChanges, step.captureIndex, castStructuredValue(content[pos:next], trimSpace, kinds[step.captureIndex]))
		}
		return next, true, nextChanges
	case len(step.alternatives) > 0:
		for _, alt := range matchingAlternatives(step, content, pos) {
			mark := len(changes)
			next, ok, nextChanges := alt.matchTypedAnyFrom(dst, content, pos, trimSpace, kinds, changes)
			if ok {
				if step.captureIndex >= 0 {
					nextChanges = appendTypedMatchChange(dst, nextChanges, step.captureIndex, castStructuredValue(content[pos:next], trimSpace, kinds[step.captureIndex]))
				}
				return next, true, nextChanges
			}
			if alt.writes {
				rollbackTypedMatchChanges(dst, nextChanges, mark)
				nextChanges = nextChanges[:mark]
			}
		}
		return 0, false, changes
	default:
		return pos, true, changes
	}
}

func (m structuredMatcher) matchTypedAnyFrom(dst []any, content string, pos int, trimSpace bool, kinds []valueKind, changes []typedMatchChange) (int, bool, []typedMatchChange) {
	if m.quickReject(content, pos) {
		return 0, false, changes
	}
	if !m.backtracking && !m.changeLog && changes == nil {
		next, ok := m.matchTypedLinearFrom(dst, content, pos, trimSpace, kinds)
		return next, ok, nil
	}
	if m.backtracking {
		return m.matchTypedBacktrackingFrom(dst, content, pos, trimSpace, kinds, changes)
	}
	return m.matchTypedFrom(dst, content, pos, trimSpace, kinds, changes)
}

func matchingAlternatives(step structuredStep, content string, pos int) []*structuredMatcher {
	if len(step.altPrefixes) == 0 {
		return step.alternatives
	}
	for i, prefix := range step.altPrefixes {
		if strings.HasPrefix(content[pos:], prefix) {
			return step.alternatives[i : i+1]
		}
	}
	return nil
}

func (m structuredMatcher) quickReject(content string, pos int) bool {
	if len(content)-pos < m.ir.MinWidth {
		return true
	}
	if m.requiresStartWordBoundary(content, pos) {
		return true
	}
	if m.ir.FirstLiteralExact && m.ir.FirstLiteral != "" && !strings.HasPrefix(content[pos:], m.ir.FirstLiteral) {
		return true
	}
	return m.anchoredStart && requiredLiteralRejects(content[pos:], m.ir.FirstLiteral, m.required)
}

func (m structuredMatcher) requiresStartWordBoundary(content string, pos int) bool {
	if len(m.steps) == 0 || pos <= 0 {
		return false
	}
	step := m.steps[0]
	if step.optional || step.literal != "" || step.parser == nil {
		return false
	}
	switch step.parser.kind {
	case structuredMonthName:
		return isWordByte(content[pos-1])
	default:
		return false
	}
}

func appendTypedMatchChange(dst []any, changes []typedMatchChange, idx int, value any) []typedMatchChange {
	if idx < 0 {
		return changes
	}
	changes = append(changes, typedMatchChange{idx: idx, prev: dst[idx]})
	dst[idx] = value
	return changes
}

func rollbackTypedMatchChanges(dst []any, changes []typedMatchChange, mark int) {
	for i := len(changes) - 1; i >= mark; i-- {
		change := changes[i]
		dst[change.idx] = change.prev
	}
}

func castStructuredValue(raw string, trimSpace bool, kind valueKind) any {
	if trimSpace {
		raw = trimMatch(raw)
	}
	v, _ := castValue(raw, kind)
	return v
}

func appendMatchChange(dst []string, changes []matchChange, idx int, value string) []matchChange {
	if idx < 0 {
		return changes
	}
	changes = append(changes, matchChange{idx: idx, prev: dst[idx]})
	dst[idx] = value
	return changes
}

func rollbackMatchChanges(dst []string, changes []matchChange, mark int) {
	for i := len(changes) - 1; i >= mark; i-- {
		change := changes[i]
		dst[change.idx] = change.prev
	}
}

func (p *structuredParser) consume(content string, pos int) (next int, value string, ok bool) {
	if pos > len(content) {
		return 0, "", false
	}
	if p.wrapPrefix != "" || p.wrapSuffix != "" {
		return p.consumeWrapped(content, pos)
	}
	if p.kind == structuredIPOrHost {
		if seg, next, ok := sliceIPv4Literal(content, pos); ok {
			return next, seg, true
		}
		if seg, next, ok := sliceIPv6Like(content, pos); ok {
			return next, seg, true
		}
		seg, next, ok := sliceHostName(content, pos)
		if !ok || !isHostnameMatchContext(content, pos, next, seg) {
			return 0, "", false
		}
		return next, seg, true
	}
	if p.kind == structuredHostName {
		seg, next, ok := sliceHostName(content, pos)
		if !ok || !isHostnameMatchContext(content, pos, next, seg) {
			return 0, "", false
		}
		return next, seg, true
	}

	segment, next, ok := p.slice(content, pos)
	if !ok {
		return 0, "", false
	}

	switch p.kind {
	case structuredWord:
		if !isWordMatchContext(content, pos, next, segment) {
			return 0, "", false
		}
	case structuredURIPath, structuredURIPathParam:
		if segment == "" {
			return 0, "", false
		}
	case structuredHostName, structuredIPOrHost:
		if segment == "" {
			return 0, "", false
		}
	case structuredCharClass:
		if !p.allowEmpty && segment == "" {
			return 0, "", false
		}
	case structuredNotSpace:
		if segment == "" {
			return 0, "", false
		}
	case structuredNumber:
		if !isApacheNumber(segment) {
			return 0, "", false
		}
	case structuredInt:
		if !isApacheInt(segment) {
			return 0, "", false
		}
	case structuredPosInt:
		if !isApachePosInt(segment) {
			return 0, "", false
		}
	case structuredNonNegInt:
		if !isApacheNonNegInt(segment) {
			return 0, "", false
		}
	case structuredMonthName:
		if !isMonthNameValue(segment) {
			return 0, "", false
		}
	case structuredDayName:
		if !isDayNameValue(segment) {
			return 0, "", false
		}
	case structuredTimeOfDay:
		if !looksLikeTimeOfDay(segment) {
			return 0, "", false
		}
	case structuredYear:
		if !isYearValue(segment) {
			return 0, "", false
		}
	case structuredMonthNum:
		if !isMonthNumValue(segment) {
			return 0, "", false
		}
	case structuredMonthDay:
		if !isMonthDayValue(segment) {
			return 0, "", false
		}
	case structuredHour:
		if !isHourValue(segment) {
			return 0, "", false
		}
	case structuredMinute:
		if !isMinuteValue(segment) {
			return 0, "", false
		}
	case structuredSecond:
		if !isSecondValue(segment) {
			return 0, "", false
		}
	case structuredQuoted:
		if len(segment) < 2 {
			return 0, "", false
		}
		q := segment[0]
		if (q != '"' && q != '\'') || segment[len(segment)-1] != q {
			return 0, "", false
		}
	case structuredUntilLiteral, structuredGreedyUntilLiteral:
		if !dotRunSegmentOK(p, segment) {
			return 0, "", false
		}
	case structuredSpaceOne:
		if len(segment) != 1 || !isOnlyRegexpSpace(segment) {
			return 0, "", false
		}
	case structuredSpacePlus:
		if segment == "" || !isOnlyRegexpSpace(segment) {
			return 0, "", false
		}
	case structuredSpaceStar:
		if !isOnlyRegexpSpace(segment) {
			return 0, "", false
		}
	case structuredTimestampISO8601:
		if segment == "" {
			return 0, "", false
		}
	case structuredHTTPDate:
		if segment == "" {
			return 0, "", false
		}
	case structuredLogLevel:
		if !isLogLevel(segment) {
			return 0, "", false
		}
	default:
		return 0, "", false
	}

	return next, segment, true
}

func (p *structuredParser) consumeWrapped(content string, pos int) (next int, value string, ok bool) {
	if p.inner == nil {
		return 0, "", false
	}

	if p.wrapPrefix != "" {
		if !strings.HasPrefix(content[pos:], p.wrapPrefix) {
			return 0, "", false
		}
		pos += len(p.wrapPrefix)
	}

	next, value, ok = p.inner.consume(content, pos)
	if !ok {
		return 0, "", false
	}
	if p.wrapSuffix != "" {
		if !strings.HasPrefix(content[next:], p.wrapSuffix) {
			return 0, "", false
		}
		next += len(p.wrapSuffix)
	}
	return next, value, true
}

func (p *structuredParser) slice(content string, pos int) (segment string, next int, ok bool) {
	switch p.kind {
	case structuredSpaceOne, structuredSpacePlus, structuredSpaceStar:
		i := pos
		for i < len(content) && isRegexpASCIISpace(content[i]) {
			i++
		}
		if p.kind == structuredSpaceOne {
			if i == pos {
				return "", 0, false
			}
			return content[pos : pos+1], pos + 1, true
		}
		if p.kind == structuredSpacePlus && i == pos {
			return "", 0, false
		}
		return content[pos:i], i, true
	}

	if p.hasNextParser {
		switch p.kind {
		case structuredWord:
			return sliceWord(content, pos)
		case structuredHostName:
			return sliceHostName(content, pos)
		case structuredIPOrHost:
			return sliceIPOrHost(content, pos)
		case structuredURIPath:
			return sliceURIPath(content, pos)
		case structuredURIPathParam:
			return sliceURIPathParam(content, pos)
		case structuredCharClass:
			if p.singleChar {
				return sliceSingleCharClass(content, pos, p.charClass)
			}
			return sliceCharClass(content, pos, p.charClass, p.allowEmpty)
		case structuredNotSpace:
			return sliceNotSpaceWithContext(content, pos, p.nextParser, p.nextLiteral)
		case structuredNumber, structuredInt, structuredPosInt, structuredNonNegInt:
			return sliceNumericWithContext(content, pos, p.nextParser, p.nextLiteral, p.kind)
		case structuredMonthName, structuredDayName:
			return sliceAlphaToken(content, pos)
		case structuredTimeOfDay:
			return sliceTimeOfDayWithContext(content, pos)
		case structuredYear:
			return sliceYear(content, pos)
		case structuredMonthNum, structuredMonthDay, structuredHour, structuredMinute, structuredSecond:
			return sliceDigitsAndPunct(content, pos)
		case structuredQuoted:
			return sliceQuoted(content, pos)
		case structuredTimestampISO8601:
			return sliceTimestampISO8601(content, pos)
		case structuredHTTPDate:
			return sliceHTTPDate(content, pos)
		case structuredLogLevel:
			return sliceLogLevelWithContext(content, pos, p.nextParser, p.nextLiteral)
		}
	}

	if p.nextLiteral != "" {
		switch p.kind {
		case structuredUntilLiteral:
			if (p.nextParser == structuredSpaceOne || p.nextParser == structuredSpacePlus || p.nextParser == structuredSpaceStar) && p.nextLiteral != "" {
				return sliceUntilSpaceOrLiteral(content, pos, p.nextLiteral)
			}
		case structuredGreedyUntilLiteral:
			if (p.nextParser == structuredSpaceOne || p.nextParser == structuredSpacePlus || p.nextParser == structuredSpaceStar) && p.nextLiteral != "" {
				return sliceUntilSpaceOrLiteral(content, pos, p.nextLiteral)
			}
		case structuredWord:
			return sliceWord(content, pos)
		case structuredCharClass:
			return sliceCharClass(content, pos, p.charClass, p.allowEmpty)
		case structuredTimestampISO8601:
			return sliceTimestampISO8601(content, pos)
		case structuredHTTPDate:
			return sliceHTTPDate(content, pos)
		case structuredHostName:
			return sliceHostName(content, pos)
		case structuredIPOrHost:
			return sliceIPOrHost(content, pos)
		case structuredURIPath:
			return sliceURIPath(content, pos)
		case structuredURIPathParam:
			return sliceURIPathParam(content, pos)
		case structuredMonthName, structuredDayName:
			return sliceAlphaToken(content, pos)
		case structuredTimeOfDay:
			return sliceTimeOfDayWithContext(content, pos)
		case structuredYear:
			return sliceYear(content, pos)
		case structuredMonthNum, structuredMonthDay, structuredHour, structuredMinute, structuredSecond:
			return sliceDigitsAndPunct(content, pos)
		}
		rest := content[pos:]
		var idx int
		if p.kind == structuredGreedyUntilLiteral {
			idx = lastLiteralIndex(rest, p.nextLiteral)
		} else {
			idx = firstLiteralIndex(rest, p.nextLiteral)
		}
		if idx < 0 {
			return "", 0, false
		}

		return rest[:idx], pos + idx, true
	}

	switch p.kind {
	case structuredWord:
		return sliceWord(content, pos)
	case structuredHostName:
		return sliceHostName(content, pos)
	case structuredIPOrHost:
		return sliceIPOrHost(content, pos)
	case structuredURIPath:
		return sliceURIPath(content, pos)
	case structuredURIPathParam:
		return sliceURIPathParam(content, pos)
	case structuredCharClass:
		if p.singleChar {
			return sliceSingleCharClass(content, pos, p.charClass)
		}
		return sliceCharClass(content, pos, p.charClass, p.allowEmpty)
	case structuredNotSpace:
		return sliceNotSpace(content, pos)
	case structuredNumber:
		return sliceNumber(content, pos)
	case structuredInt:
		return sliceInt(content, pos)
	case structuredPosInt:
		return slicePosInt(content, pos)
	case structuredNonNegInt:
		return sliceNonNegInt(content, pos)
	case structuredMonthName, structuredDayName:
		return sliceAlphaToken(content, pos)
	case structuredTimeOfDay:
		return sliceTimeOfDay(content, pos)
	case structuredYear:
		return sliceYear(content, pos)
	case structuredMonthNum, structuredMonthDay, structuredHour, structuredMinute, structuredSecond:
		return sliceDigitsAndPunct(content, pos)
	case structuredQuoted:
		return sliceQuoted(content, pos)
	case structuredTimestampISO8601:
		return sliceTimestampISO8601(content, pos)
	case structuredHTTPDate:
		return sliceHTTPDate(content, pos)
	case structuredLogLevel:
		return sliceLogLevel(content, pos)
	case structuredUntilLiteral, structuredGreedyUntilLiteral:
		if p.singleChar {
			return sliceSingleDot(content, pos)
		}
		return sliceDotRun(content, pos, p.allowEmpty, p.dotAll)
	}

	return content[pos:], len(content), true
}

func dotRunSegmentOK(p *structuredParser, segment string) bool {
	if p == nil {
		return false
	}
	if p.singleChar {
		if segment == "" || strings.IndexByte(segment, '\n') >= 0 {
			return false
		}
		_, width := utf8.DecodeRuneInString(segment)
		return width == len(segment)
	}
	if segment == "" && !p.allowEmpty {
		return false
	}
	return p.dotAll || strings.IndexByte(segment, '\n') < 0
}

func canParserEndWithoutLiteral(kind structuredKind) bool {
	switch kind {
	case structuredWord, structuredHostName, structuredIPOrHost, structuredURIPath, structuredURIPathParam, structuredCharClass, structuredNotSpace, structuredNumber, structuredInt, structuredPosInt, structuredNonNegInt, structuredMonthName, structuredDayName, structuredTimeOfDay, structuredYear, structuredMonthNum, structuredMonthDay, structuredHour, structuredMinute, structuredSecond, structuredQuoted, structuredSpaceOne, structuredSpacePlus, structuredSpaceStar, structuredTimestampISO8601, structuredHTTPDate, structuredLogLevel:
		return true
	default:
		return false
	}
}

func sliceWord(content string, pos int) (string, int, bool) {
	start := pos
	for pos < len(content) && isWordByte(content[pos]) {
		pos++
	}
	if pos == start {
		return "", 0, false
	}
	return content[start:pos], pos, true
}

func isWordMatchContext(content string, start int, end int, segment string) bool {
	if segment == "" ||
		end-start != len(segment) ||
		!hasRegexpWordBoundary(content, start) ||
		!hasRegexpWordBoundary(content, end) {
		return false
	}
	for i := 0; i < len(segment); i++ {
		if !isWordByte(segment[i]) {
			return false
		}
	}
	return true
}

func sliceURIPath(content string, pos int) (string, int, bool) {
	if pos >= len(content) || content[pos] != '/' {
		return "", 0, false
	}
	i := pos + 1
	for i < len(content) {
		c := content[i]
		if isURIPathChar(c) {
			i++
			continue
		}
		break
	}
	return content[pos:i], i, true
}

func sliceURIPathParam(content string, pos int) (string, int, bool) {
	_, next, ok := sliceURIPath(content, pos)
	if !ok {
		return "", 0, false
	}
	i := next
	if i < len(content) && content[i] == '?' {
		i++
		for i < len(content) && isURIParamChar(content[i]) {
			i++
		}
	}
	return content[pos:i], i, true
}

func isHostnameStartByte(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9')
}

func isHostnameByte(b byte) bool {
	return isHostnameStartByte(b) || b == '-' || b == '.'
}

func sliceCharClass(content string, pos int, class *asciiCharClass, allowEmpty bool) (string, int, bool) {
	if class == nil {
		return "", 0, false
	}
	start := pos
	for pos < len(content) && class.Table[content[pos]] {
		pos++
	}
	if pos == start && !allowEmpty {
		return "", 0, false
	}
	return content[start:pos], pos, true
}

func sliceSingleCharClass(content string, pos int, class *asciiCharClass) (string, int, bool) {
	if class == nil || pos >= len(content) || !class.Table[content[pos]] {
		return "", 0, false
	}
	next := pos + 1
	if content[pos] >= utf8.RuneSelf {
		_, width := utf8.DecodeRuneInString(content[pos:])
		if width > 1 {
			next = pos + width
		}
	}
	return content[pos:next], next, true
}

func sliceSingleDot(content string, pos int) (string, int, bool) {
	if pos >= len(content) || content[pos] == '\n' {
		return "", 0, false
	}
	next := pos + 1
	if content[pos] >= utf8.RuneSelf {
		_, width := utf8.DecodeRuneInString(content[pos:])
		if width > 1 {
			next = pos + width
		}
	}
	return content[pos:next], next, true
}

func sliceNotSpace(content string, pos int) (string, int, bool) {
	start := pos
	for pos < len(content) && !isRegexpASCIISpace(content[pos]) {
		pos++
	}
	if pos == start {
		return "", 0, false
	}
	return content[start:pos], pos, true
}

func sliceHostName(content string, pos int) (string, int, bool) {
	start := pos
	if pos >= len(content) || !isHostnameStartByte(content[pos]) {
		return "", 0, false
	}
	for {
		labelStart := pos
		for pos < len(content) && (isHostnameStartByte(content[pos]) || content[pos] == '-') {
			pos++
		}
		if pos-labelStart > 63 {
			if labelStart == start {
				return "", 0, false
			}
			return content[start:labelStart], labelStart, true
		}
		if pos >= len(content) || content[pos] != '.' {
			return content[start:pos], pos, true
		}
		if pos+1 >= len(content) || !isHostnameStartByte(content[pos+1]) {
			return content[start : pos+1], pos + 1, true
		}
		pos++
	}
}

func sliceIPOrHost(content string, pos int) (string, int, bool) {
	if seg, next, ok := sliceIPv4Literal(content, pos); ok {
		return seg, next, true
	}
	if seg, next, ok := sliceIPv6Like(content, pos); ok {
		return seg, next, true
	}
	return sliceHostName(content, pos)
}

func sliceIPv4Literal(content string, pos int) (string, int, bool) {
	start := pos
	for part := 0; part < 4; part++ {
		if pos >= len(content) || !isASCIIDigit(content[pos]) {
			return "", 0, false
		}

		val := 0
		digits := 0
		for pos < len(content) && isASCIIDigit(content[pos]) {
			val = val*10 + int(content[pos]-'0')
			digits++
			if digits > 3 || val > 255 {
				return "", 0, false
			}
			pos++
		}
		if part == 3 {
			break
		}
		if pos >= len(content) || content[pos] != '.' {
			return "", 0, false
		}
		pos++
	}

	if pos < len(content) && (isHostnameByte(content[pos]) || content[pos] == ':') {
		return "", 0, false
	}
	return content[start:pos], pos, true
}

func sliceNotSpaceWithContext(content string, pos int, nextParser structuredKind, nextLiteral string) (string, int, bool) {
	if nextLiteral != "" {
		requireSpace := nextParser == structuredSpaceOne || nextParser == structuredSpacePlus
		return sliceTokenBeforeSpaceOrLiteral(content, pos, nextLiteral, requireSpace)
	}
	return sliceNotSpace(content, pos)
}

func sliceNumber(content string, pos int) (string, int, bool) {
	start := pos
	signed := false
	if pos < len(content) && (content[pos] == '+' || content[pos] == '-') {
		signed = true
		pos++
	}
	digitsStart := pos
	for pos < len(content) && content[pos] >= '0' && content[pos] <= '9' {
		pos++
	}
	if pos == digitsStart {
		if signed || pos >= len(content) || content[pos] != '.' {
			return "", 0, false
		}
		pos++
		fracStart := pos
		for pos < len(content) && content[pos] >= '0' && content[pos] <= '9' {
			pos++
		}
		if pos == fracStart {
			return "", 0, false
		}
		return content[start:pos], pos, true
	}
	if pos < len(content) && content[pos] == '.' && pos+1 < len(content) && content[pos+1] >= '0' && content[pos+1] <= '9' {
		pos += 2
		for pos < len(content) && content[pos] >= '0' && content[pos] <= '9' {
			pos++
		}
	}
	return content[start:pos], pos, true
}

func sliceInt(content string, pos int) (string, int, bool) {
	start := pos
	if pos < len(content) && (content[pos] == '+' || content[pos] == '-') {
		pos++
	}
	digitsStart := pos
	for pos < len(content) && content[pos] >= '0' && content[pos] <= '9' {
		pos++
	}
	if pos == digitsStart {
		return "", 0, false
	}
	return content[start:pos], pos, true
}

func slicePosInt(content string, pos int) (string, int, bool) {
	if pos >= len(content) || content[pos] < '1' || content[pos] > '9' {
		return "", 0, false
	}
	start := pos
	pos++
	for pos < len(content) && content[pos] >= '0' && content[pos] <= '9' {
		pos++
	}
	return content[start:pos], pos, true
}

func sliceNonNegInt(content string, pos int) (string, int, bool) {
	start := pos
	for pos < len(content) && content[pos] >= '0' && content[pos] <= '9' {
		pos++
	}
	if pos == start {
		return "", 0, false
	}
	return content[start:pos], pos, true
}

func sliceNumericWithContext(content string, pos int, nextParser structuredKind, nextLiteral string, kind structuredKind) (string, int, bool) {
	if nextLiteral != "" {
		requireSpace := nextParser == structuredSpaceOne || nextParser == structuredSpacePlus
		return sliceTokenBeforeSpaceOrLiteral(content, pos, nextLiteral, requireSpace)
	}
	switch kind {
	case structuredInt:
		return sliceInt(content, pos)
	case structuredPosInt:
		return slicePosInt(content, pos)
	case structuredNonNegInt:
		return sliceNonNegInt(content, pos)
	default:
		return sliceNumber(content, pos)
	}
}

func sliceQuoted(content string, pos int) (string, int, bool) {
	if pos >= len(content) || (content[pos] != '"' && content[pos] != '\'') {
		return "", 0, false
	}
	q := content[pos]
	start := pos
	pos++
	for pos < len(content) {
		if content[pos] == '\\' && pos+1 < len(content) {
			pos += 2
			continue
		}
		if content[pos] == q {
			pos++
			return content[start:pos], pos, true
		}
		pos++
	}
	return "", 0, false
}

func sliceTimestampISO8601(content string, pos int) (string, int, bool) {
	next, ok := consumeTimestampISO8601(content, pos)
	if !ok || next == pos {
		return "", 0, false
	}
	return content[pos:next], next, true
}

func sliceHTTPDate(content string, pos int) (string, int, bool) {
	next, ok := consumeHTTPDate(content, pos)
	if !ok || next == pos {
		return "", 0, false
	}
	return content[pos:next], next, true
}

func sliceLogLevel(content string, pos int) (string, int, bool) {
	start := pos
	for pos < len(content) {
		c := content[pos]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
			pos++
			continue
		}
		break
	}
	if pos == start {
		return "", 0, false
	}
	return content[start:pos], pos, true
}

func sliceLogLevelWithContext(content string, pos int, nextParser structuredKind, nextLiteral string) (string, int, bool) {
	if (nextParser == structuredSpaceOne || nextParser == structuredSpacePlus || nextParser == structuredSpaceStar) && nextLiteral != "" {
		start := pos
		for pos < len(content) {
			c := content[pos]
			if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
				pos++
				continue
			}
			break
		}
		if pos == start {
			return "", 0, false
		}
		return content[start:pos], pos, true
	}
	return sliceLogLevel(content, pos)
}

func sliceTokenBeforeSpaceOrLiteral(content string, pos int, nextLiteral string, requireSpace bool) (string, int, bool) {
	if nextLiteral == "" || pos >= len(content) {
		return "", 0, false
	}

	tokenEnd := pos
	for tokenEnd < len(content) && !isRegexpASCIISpace(content[tokenEnd]) {
		tokenEnd++
	}

	if tokenEnd == pos {
		return "", 0, false
	}

	searchEnd := tokenEnd
	if tokenEnd < len(content) {
		litPos := tokenEnd
		for litPos < len(content) && isRegexpASCIISpace(content[litPos]) {
			litPos++
		}
		searchEnd = litPos
		expected := nextLiteral
		if litPos > tokenEnd {
			for len(expected) > 0 && isRegexpASCIISpace(expected[0]) {
				expected = expected[1:]
			}
		}
		if expected != "" && strings.HasPrefix(content[litPos:], expected) {
			if requireSpace && litPos == tokenEnd {
				return "", 0, false
			}
			return content[pos:tokenEnd], tokenEnd, true
		}
	}

	if requireSpace {
		return "", 0, false
	}

	rel := lastLiteralIndex(content[pos:searchEnd], nextLiteral)
	if rel <= 0 {
		return "", 0, false
	}
	tokenEnd = pos + rel
	return content[pos:tokenEnd], tokenEnd, true
}

func sliceUntilSpaceOrLiteral(content string, pos int, nextLiteral string) (string, int, bool) {
	if nextLiteral == "" || pos > len(content) {
		return "", 0, false
	}

	first := nextLiteral[0]
	for i := pos; i < len(content); i++ {
		c := content[i]
		if c == first && strings.HasPrefix(content[i:], nextLiteral) {
			return content[pos:i], i, true
		}
		if !isRegexpASCIISpace(c) {
			continue
		}

		j := i
		for j < len(content) && isRegexpASCIISpace(content[j]) {
			j++
		}
		if strings.HasPrefix(content[j:], nextLiteral) {
			return content[pos:i], i, true
		}
	}
	return "", 0, false
}

func firstLiteralIndex(content, literal string) int {
	if literal == "" {
		return 0
	}
	if len(literal) == 1 {
		return strings.IndexByte(content, literal[0])
	}

	first := literal[0]
	search := content
	offset := 0
	for {
		idx := strings.IndexByte(search, first)
		if idx < 0 {
			return -1
		}
		idx += offset
		if strings.HasPrefix(content[idx:], literal) {
			return idx
		}
		offset = idx + 1
		search = content[offset:]
	}
}

func lastLiteralIndex(content, literal string) int {
	if literal == "" {
		return len(content)
	}
	if len(literal) == 1 {
		return strings.LastIndexByte(content, literal[0])
	}

	first := literal[0]
	end := len(content)
	for end > 0 {
		idx := strings.LastIndexByte(content[:end], first)
		if idx < 0 {
			return -1
		}
		if strings.HasPrefix(content[idx:], literal) {
			return idx
		}
		end = idx
	}
	return -1
}

func sliceYear(content string, pos int) (string, int, bool) {
	start := pos
	for pos < len(content) && content[pos] >= '0' && content[pos] <= '9' && pos-start < 4 {
		pos++
	}
	if pos-start != 2 && pos-start != 4 {
		return "", 0, false
	}
	return content[start:pos], pos, true
}

func sliceTimeOfDay(content string, pos int) (string, int, bool) {
	next, ok := consumePatternTimeOfDay(content, pos, true, false)
	if !ok || next == pos {
		return "", 0, false
	}
	return content[pos:next], next, true
}

func sliceTimeOfDayWithContext(content string, pos int) (string, int, bool) {
	next, ok := consumePatternTimeOfDay(content, pos, true, false)
	if !ok || next == pos {
		return "", 0, false
	}
	return content[pos:next], next, true
}

func sliceDigitsAndPunct(content string, pos int) (string, int, bool) {
	start := pos
	for pos < len(content) {
		c := content[pos]
		if (c >= '0' && c <= '9') || c == '.' || c == ',' {
			pos++
			continue
		}
		break
	}
	if pos == start {
		return "", 0, false
	}
	return content[start:pos], pos, true
}

func sliceDotRun(content string, pos int, allowEmpty bool, dotAll bool) (string, int, bool) {
	start := pos
	for pos < len(content) && (dotAll || content[pos] != '\n') {
		pos++
	}
	if !allowEmpty && pos == start {
		return "", 0, false
	}
	return content[start:pos], pos, true
}

func sliceIPv6Like(content string, pos int) (string, int, bool) {
	start := pos
	colons := 0
	for pos < len(content) {
		c := content[pos]
		switch {
		case c >= '0' && c <= '9':
			pos++
		case c >= 'A' && c <= 'F':
			pos++
		case c >= 'a' && c <= 'f':
			pos++
		case c == ':' || c == '.':
			if c == ':' {
				colons++
			}
			pos++
		default:
			goto done
		}
	}
done:
	if pos == start || colons < 2 {
		return "", 0, false
	}
	if pos < len(content) && content[pos] == '%' {
		zoneStart := pos
		pos++
		for pos < len(content) && content[pos] != '\n' {
			pos++
		}
		if pos == zoneStart+1 {
			pos = zoneStart
		}
	}
	segment := content[start:pos]
	ipSegment := segment
	if zone := strings.IndexByte(ipSegment, '%'); zone >= 0 {
		ipSegment = ipSegment[:zone]
	}
	if net.ParseIP(ipSegment) == nil || !strings.Contains(ipSegment, ":") {
		return "", 0, false
	}
	return segment, pos, true
}

func isURIPathChar(c byte) bool {
	switch {
	case c >= 'A' && c <= 'Z':
		return true
	case c >= 'a' && c <= 'z':
		return true
	case c >= '0' && c <= '9':
		return true
	}
	switch c {
	case '/', '$', '.', '+', '!', '*', '\'', '(', ')', '{', '}', ',', '~', ':', ';', '=', '@', '#', '%', '_', '-':
		return true
	default:
		return false
	}
}

func isURIParamChar(c byte) bool {
	if isURIPathChar(c) {
		return true
	}
	switch c {
	case '?', '&', '|', '[', ']', '<', '>':
		return true
	default:
		return false
	}
}

func isOnlyRegexpSpace(s string) bool {
	for i := 0; i < len(s); i++ {
		if !isRegexpASCIISpace(s[i]) {
			return false
		}
	}
	return true
}

func isRegexpASCIISpace(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '\f':
		return true
	default:
		return false
	}
}

func isStructuredIPLiteral(s string) bool {
	if strings.IndexByte(s, ':') >= 0 {
		return true
	}
	_, _, ok := sliceIPv4Literal(s, 0)
	return ok
}

func isHostnameMatchContext(content string, start int, end int, segment string) bool {
	if !hasRegexpWordBoundary(content, start) {
		return false
	}
	if strings.HasSuffix(segment, ".") {
		return true
	}
	if end >= len(content) || !isWordByte(content[end]) {
		return true
	}
	return hasRegexpWordBoundary(content, end)
}

func hasRegexpWordBoundary(content string, idx int) bool {
	prevWord := idx > 0 && isWordByte(content[idx-1])
	nextWord := idx < len(content) && isWordByte(content[idx])
	return prevWord != nextWord
}

func isYearValue(s string) bool {
	return len(s) == 2 || len(s) == 4
}

func isMonthNumValue(s string) bool {
	n, ok := parsePositiveInt(s)
	return ok && n >= 1 && n <= 12 && len(s) <= 2
}

func isMonthDayValue(s string) bool {
	n, ok := parsePositiveInt(s)
	return ok && n >= 1 && n <= 31 && len(s) <= 2
}

func isHourValue(s string) bool {
	n, ok := parsePositiveInt(s)
	return ok && n >= 0 && n <= 23 && len(s) <= 2
}

func isMinuteValue(s string) bool {
	n, ok := parsePositiveInt(s)
	return ok && n >= 0 && n <= 59 && len(s) == 2
}

func isSecondValue(s string) bool {
	if s == "" {
		return false
	}
	main := s
	if idx := strings.IndexAny(s, ".,:"); idx >= 0 {
		main = s[:idx]
		if idx == len(s)-1 {
			return false
		}
		for i := idx + 1; i < len(s); i++ {
			if s[i] < '0' || s[i] > '9' {
				return false
			}
		}
	}
	n, ok := parsePositiveInt(main)
	return ok && n >= 0 && n <= 60 && len(main) == 2
}

func parsePositiveInt(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, false
		}
		n = n*10 + int(s[i]-'0')
	}
	return n, true
}

func looksLikeTimestampISO8601(s string) bool {
	next, ok := consumeTimestampISO8601(s, 0)
	return ok && next == len(s)
}

func looksLikeTimeOfDay(s string) bool {
	next, ok := consumePatternTimeOfDay(s, 0, true, false)
	return ok && next == len(s)
}

func looksLikeHTTPDate(s string) bool {
	next, ok := consumeHTTPDate(s, 0)
	return ok && next == len(s)
}

func isLogLevel(s string) bool {
	return isLogLevelValue(s)
}
