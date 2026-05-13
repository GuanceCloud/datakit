//go:build amd64
// +build amd64

package filter

type amd64ProgramCompiler struct {
	info ProgramJITInfo
}

func (c amd64ProgramCompiler) Compile(program Program) programRunner {
	return compileSpecializedProgramRunner(program)
}

func (c amd64ProgramCompiler) Info() ProgramJITInfo {
	return cloneProgramJITInfo(c.info)
}

func newPlatformProgramJITCompiler() programJITCompiler {
	return amd64ProgramCompiler{
		info: ProgramJITInfo{
			Arch:             "amd64",
			Backend:          "prejit-amd64",
			Enabled:          false,
			SupportedOpcodes: supportedProgramJITOpcodes(),
		},
	}
}
