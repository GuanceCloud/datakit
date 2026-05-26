package filter

import "strings"

type Spec struct {
	AnchoredPrefix string
	LiteralPrefix  string
	LiteralExact   bool
	ExactByLen     map[int][]string
	LiteralSet     []string
	Required       []string
}

type MatcherSetFilter struct {
	anchoredPrefix  string
	literalPrefix   int
	exactByLen      map[int][]string
	anyAtomIDs      []int
	requiredAtomIDs []int
	required        []string
	program         Program
	useProgram      bool
}

type EvalContext struct {
	Content  string
	AtomBits uint64
	UseBits  bool
	AtomHits []bool
}

type opcode uint8

const (
	opPrefix opcode = iota + 1
	opExactByLen
	opHasAtom
	opHasAnyAtom
	opHasAllAtoms
	opRequiredOrder
)

type instruction struct {
	op        opcode
	intArg    int
	stringArg string
	intsArg   []int
	exactArg  map[int][]string
	litsArg   []string
}

type Program struct {
	code         []instruction
	runner       programRunner
	prefixRunner *prefixProgramRunner
	exactRunner  *exactSetProgramRunner
	atomRunner   *atomGateProgramRunner
	comboRunner  *exactAtomProgramRunner
}

func Compile(spec Spec, atomIDs map[string]int) MatcherSetFilter {
	filter := MatcherSetFilter{
		anchoredPrefix: spec.AnchoredPrefix,
		literalPrefix:  -1,
	}
	if spec.LiteralExact && len(spec.ExactByLen) > 0 {
		filter.exactByLen = cloneLiteralLengthBuckets(spec.ExactByLen)
	}
	if !spec.LiteralExact {
		filter.anyAtomIDs = compileAtomRefs(spec.LiteralSet, atomIDs, "")
	}
	filter.requiredAtomIDs = compileAtomRefs(spec.Required, atomIDs, spec.AnchoredPrefix)
	filter.required = append([]string(nil), spec.Required...)
	if spec.LiteralPrefix != "" {
		if atomID, ok := atomIDs[spec.LiteralPrefix]; ok {
			filter.literalPrefix = atomID
		}
	}
	filter.program = compileProgram(filter)
	filter.useProgram = shouldUseProgram(filter)
	return filter
}

func shouldUseProgram(filter MatcherSetFilter) bool {
	if filter.program.prefixRunner != nil {
		return false
	}
	return filter.program.RunnerKind() == "specialized"
}

func compileProgram(filter MatcherSetFilter) Program {
	code := make([]instruction, 0, 6)
	if filter.anchoredPrefix != "" {
		code = append(code, instruction{
			op:        opPrefix,
			stringArg: filter.anchoredPrefix,
		})
	}
	if len(filter.exactByLen) > 0 {
		code = append(code, instruction{
			op:       opExactByLen,
			exactArg: cloneLiteralLengthBuckets(filter.exactByLen),
		})
	}
	if filter.literalPrefix >= 0 {
		code = append(code, instruction{
			op:     opHasAtom,
			intArg: filter.literalPrefix,
		})
	}
	if len(filter.anyAtomIDs) > 0 {
		code = append(code, instruction{
			op:      opHasAnyAtom,
			intsArg: append([]int(nil), filter.anyAtomIDs...),
		})
	}
	if len(filter.requiredAtomIDs) > 0 {
		code = append(code, instruction{
			op:      opHasAllAtoms,
			intsArg: append([]int(nil), filter.requiredAtomIDs...),
		})
	}
	if len(filter.required) > 0 {
		code = append(code, instruction{
			op:        opRequiredOrder,
			stringArg: filter.anchoredPrefix,
			litsArg:   append([]string(nil), filter.required...),
		})
	}
	program := Program{code: code}
	program.runner = compileProgramRunner(program)
	switch runner := program.runner.(type) {
	case *prefixProgramRunner:
		program.prefixRunner = runner
	case *exactSetProgramRunner:
		program.exactRunner = runner
	case *atomGateProgramRunner:
		program.atomRunner = runner
	case *exactAtomProgramRunner:
		program.comboRunner = runner
	}
	return program
}

func compileAtomRefs(atoms []string, atomIDs map[string]int, skip string) []int {
	if len(atoms) == 0 || len(atomIDs) == 0 {
		return nil
	}

	seen := make(map[int]struct{}, len(atoms))
	out := make([]int, 0, len(atoms))
	for _, atom := range atoms {
		if atom == "" || atom == skip {
			continue
		}
		atomID, ok := atomIDs[atom]
		if !ok {
			continue
		}
		if _, exists := seen[atomID]; exists {
			continue
		}
		seen[atomID] = struct{}{}
		out = append(out, atomID)
	}
	return out
}

func cloneLiteralLengthBuckets(src map[int][]string) map[int][]string {
	if len(src) == 0 {
		return nil
	}
	out := make(map[int][]string, len(src))
	for length, literals := range src {
		out[length] = append([]string(nil), literals...)
	}
	return out
}

func (f MatcherSetFilter) Accepts(ctx EvalContext) bool {
	if f.useProgram && len(f.program.code) > 0 {
		return f.program.Run(ctx)
	}
	return f.RunStruct(ctx)
}

func (f MatcherSetFilter) RunStruct(ctx EvalContext) bool {
	if f.anchoredPrefix != "" && !strings.HasPrefix(ctx.Content, f.anchoredPrefix) {
		return false
	}
	if len(f.exactByLen) > 0 && !matchesExact(ctx.Content, f.exactByLen) {
		return false
	}
	if f.literalPrefix >= 0 && !ctx.HasAtom(f.literalPrefix) {
		return false
	}
	if len(f.anyAtomIDs) > 0 && !hasAny(ctx, f.anyAtomIDs) {
		return false
	}
	if len(f.requiredAtomIDs) > 0 && !hasAll(ctx, f.requiredAtomIDs) {
		return false
	}
	if requiredLiteralRejects(ctx.Content, f.anchoredPrefix, f.required) {
		return false
	}
	return true
}

func (f MatcherSetFilter) RunProgram(ctx EvalContext) bool {
	return f.program.Run(ctx)
}

func (f MatcherSetFilter) ProgramJITEnabled() bool {
	return f.program.JITEnabled()
}

func (f MatcherSetFilter) ProgramJITCandidate() bool {
	return f.program.JITCandidate()
}

func (f MatcherSetFilter) ProgramJITInfo() ProgramJITInfo {
	return f.program.JITInfo()
}

func (f MatcherSetFilter) ProgramRunnerKind() string {
	return f.program.RunnerKind()
}

func (p Program) Run(ctx EvalContext) bool {
	if p.prefixRunner != nil {
		return p.prefixRunner.Run(ctx)
	}
	if p.exactRunner != nil {
		return p.exactRunner.Run(ctx)
	}
	if p.atomRunner != nil {
		return p.atomRunner.Run(ctx)
	}
	if p.comboRunner != nil {
		return p.comboRunner.Run(ctx)
	}
	if p.runner != nil {
		return p.runner.Run(ctx)
	}
	return p.runInterpreted(ctx)
}

func (p Program) JITEnabled() bool {
	return p.runner != nil && p.runner.Kind() == programRunnerKindJIT
}

func (p Program) RunnerKind() string {
	if p.runner == nil {
		return ""
	}
	switch p.runner.Kind() {
	case programRunnerKindInterpreter:
		return "interpreter"
	case programRunnerKindSpecialized:
		return "specialized"
	case programRunnerKindJIT:
		return "jit"
	default:
		return ""
	}
}

func (p Program) JITCandidate() bool {
	return isProgramJITCandidate(p)
}

func (p Program) JITInfo() ProgramJITInfo {
	return programJITInfo()
}

func (p Program) runInterpreted(ctx EvalContext) bool {
	for _, ins := range p.code {
		switch ins.op {
		case opPrefix:
			if !strings.HasPrefix(ctx.Content, ins.stringArg) {
				return false
			}
		case opExactByLen:
			if !matchesExact(ctx.Content, ins.exactArg) {
				return false
			}
		case opHasAtom:
			if !ctx.HasAtom(ins.intArg) {
				return false
			}
		case opHasAnyAtom:
			if !hasAny(ctx, ins.intsArg) {
				return false
			}
		case opHasAllAtoms:
			if !hasAll(ctx, ins.intsArg) {
				return false
			}
		case opRequiredOrder:
			if requiredLiteralRejects(ctx.Content, ins.stringArg, ins.litsArg) {
				return false
			}
		}
	}
	return true
}

func matchesExact(content string, buckets map[int][]string) bool {
	for _, lit := range buckets[len(content)] {
		if content == lit {
			return true
		}
	}
	return false
}

func hasAny(ctx EvalContext, ids []int) bool {
	for _, id := range ids {
		if ctx.HasAtom(id) {
			return true
		}
	}
	return false
}

func hasAll(ctx EvalContext, ids []int) bool {
	for _, id := range ids {
		if !ctx.HasAtom(id) {
			return false
		}
	}
	return true
}

func (ctx EvalContext) HasAtom(id int) bool {
	if id < 0 {
		return false
	}
	if ctx.UseBits {
		if id >= 64 {
			return false
		}
		return ctx.AtomBits&(uint64(1)<<id) != 0
	}
	return id < len(ctx.AtomHits) && ctx.AtomHits[id]
}

func containsLiteralsInOrder(content string, literals []string) bool {
	for _, lit := range literals {
		if lit == "" {
			continue
		}
		idx := strings.Index(content, lit)
		if idx < 0 {
			return false
		}
		content = content[idx+len(lit):]
	}
	return true
}

func requiredLiteralRejects(content string, anchoredPrefix string, required []string) bool {
	if len(required) == 0 {
		return false
	}
	if anchoredPrefix != "" && required[0] == anchoredPrefix {
		if !strings.HasPrefix(content, anchoredPrefix) {
			return true
		}
		content = content[len(anchoredPrefix):]
		required = required[1:]
	}
	return !containsLiteralsInOrder(content, required)
}
