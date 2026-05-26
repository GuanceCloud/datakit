package filter

type programRunnerKind uint8

const (
	programRunnerKindInterpreter programRunnerKind = iota + 1
	programRunnerKindSpecialized
	programRunnerKindJIT
)

type programRunner interface {
	Run(EvalContext) bool
	Kind() programRunnerKind
}

type ProgramJITInfo struct {
	Arch             string
	Backend          string
	Enabled          bool
	SupportedOpcodes []opcode
}

type programInterpreter struct {
	program Program
}

func (r programInterpreter) Run(ctx EvalContext) bool {
	return r.program.runInterpreted(ctx)
}

func (programInterpreter) Kind() programRunnerKind {
	return programRunnerKindInterpreter
}

type prefixProgramRunner struct {
	prefix string
}

func (r *prefixProgramRunner) Run(ctx EvalContext) bool {
	return archHasPrefix(ctx.Content, r.prefix)
}

func (*prefixProgramRunner) Kind() programRunnerKind {
	return programRunnerKindSpecialized
}

type exactSetProgramRunner struct {
	prefix     string
	exactLen   int
	exactLits  []string
	exactByLen map[int][]string
}

func (r *exactSetProgramRunner) Run(ctx EvalContext) bool {
	if r.prefix != "" && !archHasPrefix(ctx.Content, r.prefix) {
		return false
	}
	if r.exactLen > 0 {
		if len(ctx.Content) != r.exactLen {
			return false
		}
		switch len(r.exactLits) {
		case 0:
			return false
		case 1:
			return archStringEqual(ctx.Content, r.exactLits[0])
		case 2:
			return archStringEqual(ctx.Content, r.exactLits[0]) || archStringEqual(ctx.Content, r.exactLits[1])
		case 3:
			return archStringEqual(ctx.Content, r.exactLits[0]) || archStringEqual(ctx.Content, r.exactLits[1]) || archStringEqual(ctx.Content, r.exactLits[2])
		default:
			for _, lit := range r.exactLits {
				if archStringEqual(ctx.Content, lit) {
					return true
				}
			}
			return false
		}
	}
	return matchesExact(ctx.Content, r.exactByLen)
}

func (r *exactSetProgramRunner) runNoPrefix(ctx EvalContext) bool {
	if r.exactLen > 0 {
		if len(ctx.Content) != r.exactLen {
			return false
		}
		switch len(r.exactLits) {
		case 0:
			return false
		case 1:
			return archStringEqual(ctx.Content, r.exactLits[0])
		case 2:
			return archStringEqual(ctx.Content, r.exactLits[0]) || archStringEqual(ctx.Content, r.exactLits[1])
		case 3:
			return archStringEqual(ctx.Content, r.exactLits[0]) || archStringEqual(ctx.Content, r.exactLits[1]) || archStringEqual(ctx.Content, r.exactLits[2])
		default:
			for _, lit := range r.exactLits {
				if archStringEqual(ctx.Content, lit) {
					return true
				}
			}
			return false
		}
	}
	return matchesExact(ctx.Content, r.exactByLen)
}

func (*exactSetProgramRunner) Kind() programRunnerKind {
	return programRunnerKindSpecialized
}

type atomGateProgramRunner struct {
	prefix         string
	hasAtomID      int
	hasAnyAtomID   []int
	hasAllAtomID   []int
	hasAtomBit     uint64
	hasAnyAtomBits uint64
	hasAllAtomBits uint64
}

func (r *atomGateProgramRunner) Run(ctx EvalContext) bool {
	if r.prefix != "" && !archHasPrefix(ctx.Content, r.prefix) {
		return false
	}
	if ctx.UseBits {
		if r.hasAtomBit != 0 && ctx.AtomBits&r.hasAtomBit == 0 {
			return false
		}
		if r.hasAnyAtomBits != 0 && ctx.AtomBits&r.hasAnyAtomBits == 0 {
			return false
		}
		if r.hasAllAtomBits != 0 && ctx.AtomBits&r.hasAllAtomBits != r.hasAllAtomBits {
			return false
		}
	} else {
		if r.hasAtomID >= 0 && !ctx.HasAtom(r.hasAtomID) {
			return false
		}
		if len(r.hasAnyAtomID) > 0 && !hasAny(ctx, r.hasAnyAtomID) {
			return false
		}
		if len(r.hasAllAtomID) > 0 && !hasAll(ctx, r.hasAllAtomID) {
			return false
		}
	}
	return true
}

func (r *atomGateProgramRunner) runNoPrefix(ctx EvalContext) bool {
	if ctx.UseBits {
		if r.hasAtomBit != 0 && ctx.AtomBits&r.hasAtomBit == 0 {
			return false
		}
		if r.hasAnyAtomBits != 0 && ctx.AtomBits&r.hasAnyAtomBits == 0 {
			return false
		}
		if r.hasAllAtomBits != 0 && ctx.AtomBits&r.hasAllAtomBits != r.hasAllAtomBits {
			return false
		}
	} else {
		if r.hasAtomID >= 0 && !ctx.HasAtom(r.hasAtomID) {
			return false
		}
		if len(r.hasAnyAtomID) > 0 && !hasAny(ctx, r.hasAnyAtomID) {
			return false
		}
		if len(r.hasAllAtomID) > 0 && !hasAll(ctx, r.hasAllAtomID) {
			return false
		}
	}
	return true
}

func (*atomGateProgramRunner) Kind() programRunnerKind {
	return programRunnerKindSpecialized
}

type exactAtomProgramRunner struct {
	prefix         string
	exactLen       int
	exactLits      []string
	exactByLen     map[int][]string
	hasAtomID      int
	hasAnyAtomID   []int
	hasAllAtomID   []int
	hasAtomBit     uint64
	hasAnyAtomBits uint64
	hasAllAtomBits uint64
}

func (r *exactAtomProgramRunner) Run(ctx EvalContext) bool {
	if r.prefix != "" && !archHasPrefix(ctx.Content, r.prefix) {
		return false
	}
	if r.exactLen > 0 {
		if len(ctx.Content) != r.exactLen {
			return false
		}
		switch len(r.exactLits) {
		case 0:
			return false
		case 1:
			if !archStringEqual(ctx.Content, r.exactLits[0]) {
				return false
			}
		case 2:
			if !archStringEqual(ctx.Content, r.exactLits[0]) && !archStringEqual(ctx.Content, r.exactLits[1]) {
				return false
			}
		case 3:
			if !archStringEqual(ctx.Content, r.exactLits[0]) && !archStringEqual(ctx.Content, r.exactLits[1]) && !archStringEqual(ctx.Content, r.exactLits[2]) {
				return false
			}
		default:
			matched := false
			for _, lit := range r.exactLits {
				if archStringEqual(ctx.Content, lit) {
					matched = true
					break
				}
			}
			if !matched {
				return false
			}
		}
	} else if len(r.exactByLen) > 0 && !matchesExact(ctx.Content, r.exactByLen) {
		return false
	}

	if ctx.UseBits {
		if r.hasAtomBit != 0 && ctx.AtomBits&r.hasAtomBit == 0 {
			return false
		}
		if r.hasAnyAtomBits != 0 && ctx.AtomBits&r.hasAnyAtomBits == 0 {
			return false
		}
		if r.hasAllAtomBits != 0 && ctx.AtomBits&r.hasAllAtomBits != r.hasAllAtomBits {
			return false
		}
	} else {
		if r.hasAtomID >= 0 && !ctx.HasAtom(r.hasAtomID) {
			return false
		}
		if len(r.hasAnyAtomID) > 0 && !hasAny(ctx, r.hasAnyAtomID) {
			return false
		}
		if len(r.hasAllAtomID) > 0 && !hasAll(ctx, r.hasAllAtomID) {
			return false
		}
	}
	return true
}

func (*exactAtomProgramRunner) Kind() programRunnerKind {
	return programRunnerKindSpecialized
}

type programJITCompiler interface {
	Compile(Program) programRunner
	Info() ProgramJITInfo
}

type disabledProgramJITCompiler struct {
	info ProgramJITInfo
}

func (disabledProgramJITCompiler) Compile(Program) programRunner {
	return nil
}

func (c disabledProgramJITCompiler) Info() ProgramJITInfo {
	return cloneProgramJITInfo(c.info)
}

var defaultProgramJITCompiler programJITCompiler = newPlatformProgramJITCompiler()

func compileProgramRunner(program Program) programRunner {
	if isProgramJITCandidate(program) {
		if runner := defaultProgramJITCompiler.Compile(program); runner != nil {
			return runner
		}
	}
	if len(program.code) == 0 {
		return nil
	}
	return programInterpreter{program: program}
}

func isProgramJITCandidate(program Program) bool {
	if len(program.code) == 0 {
		return false
	}

	info := defaultProgramJITCompiler.Info()
	if len(info.SupportedOpcodes) == 0 {
		return false
	}

	allowed := make(map[opcode]struct{}, len(info.SupportedOpcodes))
	for _, op := range info.SupportedOpcodes {
		allowed[op] = struct{}{}
	}
	for _, ins := range program.code {
		if _, ok := allowed[ins.op]; !ok {
			return false
		}
	}
	return true
}

func supportedProgramJITOpcodes() []opcode {
	return []opcode{
		opPrefix,
		opExactByLen,
		opHasAtom,
		opHasAnyAtom,
		opHasAllAtoms,
	}
}

func cloneProgramJITInfo(info ProgramJITInfo) ProgramJITInfo {
	if len(info.SupportedOpcodes) > 0 {
		info.SupportedOpcodes = append([]opcode(nil), info.SupportedOpcodes...)
	}
	return info
}

func programJITInfo() ProgramJITInfo {
	return cloneProgramJITInfo(defaultProgramJITCompiler.Info())
}

func newDisabledProgramJITCompiler(info ProgramJITInfo) programJITCompiler {
	info.Enabled = false
	return disabledProgramJITCompiler{info: info}
}

func compileSpecializedProgramRunner(program Program) programRunner {
	if len(program.code) == 0 {
		return nil
	}

	prefix := ""
	exact := &exactSetProgramRunner{}
	atom := &atomGateProgramRunner{hasAtomID: -1}
	hasExact := false
	hasAtomGate := false
	for _, ins := range program.code {
		switch ins.op {
		case opPrefix:
			prefix = ins.stringArg
			exact.prefix = ins.stringArg
			atom.prefix = ins.stringArg
		case opExactByLen:
			hasExact = true
			if len(ins.exactArg) == 1 {
				for exactLen, lits := range ins.exactArg {
					exact.exactLen = exactLen
					exact.exactLits = append([]string(nil), lits...)
				}
			} else {
				exact.exactByLen = cloneLiteralLengthBuckets(ins.exactArg)
			}
		case opHasAtom:
			hasAtomGate = true
			atom.hasAtomID = ins.intArg
			if ins.intArg >= 0 && ins.intArg < 64 {
				atom.hasAtomBit = uint64(1) << ins.intArg
			}
		case opHasAnyAtom:
			hasAtomGate = true
			atom.hasAnyAtomID = append([]int(nil), ins.intsArg...)
			atom.hasAnyAtomBits = atomBitMask(ins.intsArg)
		case opHasAllAtoms:
			hasAtomGate = true
			atom.hasAllAtomID = append([]int(nil), ins.intsArg...)
			atom.hasAllAtomBits = atomBitMask(ins.intsArg)
		default:
			return nil
		}
	}
	switch {
	case prefix != "" && !hasExact && !hasAtomGate:
		return &prefixProgramRunner{prefix: prefix}
	case hasExact && !hasAtomGate:
		return exact
	case !hasExact && hasAtomGate:
		return atom
	case hasExact && hasAtomGate:
		return &exactAtomProgramRunner{
			prefix:         exact.prefix,
			exactLen:       exact.exactLen,
			exactLits:      append([]string(nil), exact.exactLits...),
			exactByLen:     cloneLiteralLengthBuckets(exact.exactByLen),
			hasAtomID:      atom.hasAtomID,
			hasAnyAtomID:   append([]int(nil), atom.hasAnyAtomID...),
			hasAllAtomID:   append([]int(nil), atom.hasAllAtomID...),
			hasAtomBit:     atom.hasAtomBit,
			hasAnyAtomBits: atom.hasAnyAtomBits,
			hasAllAtomBits: atom.hasAllAtomBits,
		}
	default:
		return nil
	}
}

func atomBitMask(ids []int) uint64 {
	var mask uint64
	for _, id := range ids {
		if id < 0 || id >= 64 {
			return 0
		}
		mask |= uint64(1) << id
	}
	return mask
}
