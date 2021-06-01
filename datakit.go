package datakit

import (
	"os"
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

	Docker  = false
	AllOS   = []string{OSWindows, OSLinux, OSDarwin}
	AllArch = []string{OSArchWinAmd64, OSArchWin386, OSArchLinuxArm, OSArchLinuxArm64, OSArchLinux386, OSArchLinuxAmd64, OSArchDarwinAmd64}

	UnknownOS   = []string{"unknown"}
	UnknownArch = []string{"unknown"}

	OutputFile = ""

	InstallDir = optionalInstallDir[runtime.GOOS+"/"+runtime.GOARCH]

	DataDir  = filepath.Join(InstallDir, "data")
	ConfdDir = filepath.Join(InstallDir, "conf.d")

	MainConfPathDeprecated = filepath.Join(InstallDir, "datakit.conf")
	MainConfPath           = filepath.Join(ConfdDir, "datakit.conf")

	l                  = logger.DefaultSLogger("datakit")
	PipelineDir        = filepath.Join(InstallDir, "pipeline")
	PipelinePatternDir = filepath.Join(PipelineDir, "pattern")
	GRPCDomainSock     = filepath.Join(InstallDir, "datakit.sock")
	GRPCSock           = ""

	optionalInstallDir = map[string]string{
		OSArchWinAmd64: filepath.Join(`C:\Program Files`, ServiceName),
		OSArchWin386:   filepath.Join(`C:\Program Files (x86)`, ServiceName),

		OSArchLinuxArm:    filepath.Join(`/usr/local/`, ServiceName),
		OSArchLinuxArm64:  filepath.Join(`/usr/local/`, ServiceName),
		OSArchLinuxAmd64:  filepath.Join(`/usr/local/`, ServiceName),
		OSArchLinux386:    filepath.Join(`/usr/local/`, ServiceName),
		OSArchDarwinAmd64: filepath.Join(`/usr/local/`, ServiceName),
	}
)

func Quit() {
	Exit.Close()
	WG.Wait()
	close(waitstopCh)
}

func CreateSymlinks() error {

	x := [][2]string{}

	if runtime.GOOS == OSWindows {
		x = [][2]string{
			[2]string{
				filepath.Join(InstallDir, "datakit.exe"),
				`C:\WINDOWS\system32\datakit.exe`,
			},
		}
	} else {
		x = [][2]string{
			[2]string{
				filepath.Join(InstallDir, "datakit"),
				"/usr/local/bin/datakit",
			},
		}
	}

	for _, item := range x {
		if err := symlink(item[0], item[1]); err != nil {
			return err
		}
	}

	return nil
}

func symlink(src, dst string) error {

	l.Debugf("remove link %s...", dst)
	if err := os.Remove(dst); err != nil {
		l.Warnf("%s, ignored", err)
	}

	if err := os.Symlink(src, dst); err != nil {
		l.Errorf("create datakit soft link: %s -> %s: %s", dst, src, err.Error())
		return err
	}
	return nil
}
