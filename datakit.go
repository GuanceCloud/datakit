package datakit

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
)

var (
	l *logger.Logger

	Exit *cliutils.Sem
	WG   sync.WaitGroup = sync.WaitGroup{}

	DKUserAgent = fmt.Sprintf("datakit(%s), %s-%s", git.Version, runtime.GOOS, runtime.GOARCH)

	ServiceName = "datakit"

	AgentLogFile string

	MaxLifeCheckInterval time.Duration

	InstallDir     = ""
	TelegrafDir    = ""
	DataDir        = ""
	LuaDir         = ""
	ConfdDir       = ""
	GRPCDomainSock = ""

	OutputFile = ""
)

func Init() {
	l = logger.SLogger("datakit")
}

func MonitProc(proc *os.Process, name string) {
	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			p, err := os.FindProcess(proc.Pid)
			if err != nil {
				l.Error(err)
				continue
			}

			switch runtime.GOOS {
			case "windows":
				l.Debugf("%s on PID %d ok", name, proc.Pid)

			default:
				if err := p.Signal(syscall.Signal(0)); err != nil {
					l.Errorf("signal 0 to %s failed: %s", name, err)
				}
			}

		case <-Exit.Wait():
			l.Infof("exit, killing %s...", name)
			if err := proc.Kill(); err != nil { // XXX: should we wait here?
				l.Warnf("killing %s failed: %s, ignored", name, err)
			}

			l.Infof("killing %s (PID: %d) ok", name, proc.Pid)
			return
		}
	}
}
