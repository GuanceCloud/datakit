package datakit

import (
	"path/filepath"
	"runtime"
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/service"
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

	// categories
	MetricDeprecated  = "/v1/write/metrics"
	Metric            = "/v1/write/metric"
	KeyEvent          = "/v1/write/keyevent"
	Object            = "/v1/write/object"
	Logging           = "/v1/write/logging"
	Tracing           = "/v1/write/tracing"
	Rum               = "/v1/write/rum"
	Security          = "/v1/write/security"
	HeartBeat         = "/v1/write/heartbeat"
	Election          = "/v1/election"
	ElectionHeartbeat = "/v1/election/heartbeat"
)

var (
	Exit = cliutils.NewSem()
	WG   = sync.WaitGroup{}

	InstallDir         = optionalInstallDir[runtime.GOOS+"/"+runtime.GOARCH]
	optionalInstallDir = map[string]string{
		OSArchWinAmd64: `C:\Program Files\datakit`,
		OSArchWin386:   `C:\Program Files (x86)\datakit`,

		OSArchLinuxArm:    `/usr/local/datakit`,
		OSArchLinuxArm64:  `/usr/local/datakit`,
		OSArchLinuxAmd64:  `/usr/local/datakit`,
		OSArchLinux386:    `/usr/local/datakit`,
		OSArchDarwinAmd64: `/usr/local/datakit`,
	}

	AllOS   = []string{OSWindows, OSLinux, OSDarwin}
	AllArch = []string{OSArchWinAmd64, OSArchWin386, OSArchLinuxArm, OSArchLinuxArm64, OSArchLinux386, OSArchLinuxAmd64, OSArchDarwinAmd64}

	UnknownOS   = []string{"unknown"}
	UnknownArch = []string{"unknown"}

	DataDir  = filepath.Join(InstallDir, "data")
	ConfdDir = filepath.Join(InstallDir, "conf.d")

	MainConfPathDeprecated = filepath.Join(InstallDir, "datakit.conf")
	MainConfPath           = filepath.Join(ConfdDir, "datakit.conf")

	PipelineDir        = filepath.Join(InstallDir, "pipeline")
	PipelinePatternDir = filepath.Join(PipelineDir, "pattern")
	GRPCDomainSock     = filepath.Join(InstallDir, "datakit.sock")
	GRPCSock           = ""
)

func Quit() {
	Exit.Close()
	WG.Wait()
	service.Stop()
}
