package datakit

import (
	"path/filepath"
	"runtime"
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

const (
	OSWindows = `windows`
	OSLinux   = `linux`
	OSDarwin  = `darwin`

	OSArchWinAmd64    = "windows/amd64"
	OSArchWin386      = "windows/386"
	OSArchLinuxArm    = "linux/arm"
	OSArchLinuxArm64  = "linux/arm64"
	OSArchLinux386    = "linux/386"
	OSArchLinuxAmd64  = "linux/amd64"
	OSArchDarwinAmd64 = "darwin/amd64"

	CommonChanCap = 32
)

var (
	Exit = cliutils.NewSem()
	WG   = sync.WaitGroup{}

	Docker  = false
	AllOS   = []string{OSWindows, OSLinux, OSDarwin}
	AllArch = []string{OSArchWinAmd64, OSArchWin386, OSArchLinuxArm, OSArchLinuxArm64, OSArchLinux386, OSArchLinuxAmd64, OSArchDarwinAmd64}

	UnknownOS   = []string{"unknown"}
	UnknownArch = []string{"unknown"}

	OutputFile = ""

	InstallDir = optionalInstallDir[runtime.GOOS+"/"+runtime.GOARCH]

	UUIDFile = filepath.Join(InstallDir, ".id")
	DataDir  = filepath.Join(InstallDir, "data")
	LuaDir   = filepath.Join(InstallDir, "lua")
	ConfdDir = filepath.Join(InstallDir, "conf.d")

	MainConfPathDeprecated = filepath.Join(InstallDir, "datakit.conf")
	MainConfPath           = filepath.Join(ConfdDir, "datakit.conf")

	PipelineDir        = filepath.Join(InstallDir, "pipeline")
	PipelinePatternDir = filepath.Join(PipelineDir, "pattern")
	GRPCDomainSock     = filepath.Join(InstallDir, "datakit.sock")
	GRPCSock           = ""

	l                  = logger.DefaultSLogger("datakit")
	optionalInstallDir = map[string]string{
		OSArchWinAmd64: filepath.Join(`C:\Program Files\dataflux\` + ServiceName),
		OSArchWin386:   filepath.Join(`C:\Program Files (x86)\dataflux\` + ServiceName),

		OSArchLinuxArm:    filepath.Join(`/usr/local/cloudcare/dataflux/`, ServiceName),
		OSArchLinuxArm64:  filepath.Join(`/usr/local/cloudcare/dataflux/`, ServiceName),
		OSArchLinuxAmd64:  filepath.Join(`/usr/local/cloudcare/dataflux/`, ServiceName),
		OSArchLinux386:    filepath.Join(`/usr/local/cloudcare/dataflux/`, ServiceName),
		OSArchDarwinAmd64: filepath.Join(`/usr/local/cloudcare/dataflux/`, ServiceName),
	}
)

func Quit() {
	Exit.Close()
	WG.Wait()
	close(waitstopCh)
}
