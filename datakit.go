package datakit

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
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

	ReleaseCheckedInputs = "checked"
	ReleaseAllInputs     = "all"
)

var (
	ReleaseType = "" // default only release checked inputs

	Exit = cliutils.NewSem()
	WG   = sync.WaitGroup{}
	l    = logger.DefaultSLogger("datakit")

	GlobalExit = cliutils.NewSem()
	GlobalWG   = sync.WaitGroup{}

	DKUserAgent = fmt.Sprintf("datakit(%s), %s-%s", git.Version, runtime.GOOS, runtime.GOARCH)

	MaxLifeCheckInterval time.Duration

	Docker = false

	OutputFile = ""

	optionalInstallDir = map[string]string{
		OSArchWinAmd64: filepath.Join(`C:\Program Files\dataflux\` + ServiceName),
		OSArchWin386:   filepath.Join(`C:\Program Files (x86)\dataflux\` + ServiceName),

		OSArchLinuxArm:    filepath.Join(`/usr/local/cloudcare/dataflux/`, ServiceName),
		OSArchLinuxArm64:  filepath.Join(`/usr/local/cloudcare/dataflux/`, ServiceName),
		OSArchLinuxAmd64:  filepath.Join(`/usr/local/cloudcare/dataflux/`, ServiceName),
		OSArchLinux386:    filepath.Join(`/usr/local/cloudcare/dataflux/`, ServiceName),
		OSArchDarwinAmd64: filepath.Join(`/usr/local/cloudcare/dataflux/`, ServiceName),
	}

	InstallDir = optionalInstallDir[runtime.GOOS+"/"+runtime.GOARCH]

	AgentLogFile   = filepath.Join(InstallDir, "embed", "agent.log")
	TelegrafDir    = filepath.Join(InstallDir, "embed")
	DataDir        = filepath.Join(InstallDir, "data")
	LuaDir         = filepath.Join(InstallDir, "lua")
	MainConfPath   = filepath.Join(InstallDir, "datakit.conf")
	ConfdDir       = filepath.Join(InstallDir, "conf.d")
	GRPCDomainSock = filepath.Join(InstallDir, "datakit.sock")
)
