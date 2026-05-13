//go:build arm64
// +build arm64

package filter

type arm64ProgramCompiler struct {
	info ProgramJITInfo
}

func (c arm64ProgramCompiler) Compile(program Program) programRunner {
	return compileSpecializedProgramRunner(program)
}

func (c arm64ProgramCompiler) Info() ProgramJITInfo {
	return cloneProgramJITInfo(c.info)
}

func newPlatformProgramJITCompiler() programJITCompiler {
	return arm64ProgramCompiler{
		info: ProgramJITInfo{
			Arch:             "arm64",
			Backend:          "prejit-arm64",
			Enabled:          false,
			SupportedOpcodes: supportedProgramJITOpcodes(),
		},
	}
}
