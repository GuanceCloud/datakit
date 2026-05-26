//go:build !amd64 && !arm64
// +build !amd64,!arm64

package filter

func newPlatformProgramJITCompiler() programJITCompiler {
	return newDisabledProgramJITCompiler(ProgramJITInfo{
		Arch:    "generic",
		Backend: "stub-generic",
	})
}
