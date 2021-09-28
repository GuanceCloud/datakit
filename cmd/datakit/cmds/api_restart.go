package cmds

import (
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/kardianos/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkservice "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
)

var program = filepath.Join(datakit.InstallDir, getBinName())

func apiRestart() {
	goos := runtime.GOOS
	// linux: service restart
	if goos == "linux" {
		svc, err := dkservice.NewService()
		if err != nil {
			l.Error(err)
			return
		}

		// ignore signal terminated in linux
		_ = service.Control(svc, "restart")

		return
	}

	tick := time.NewTicker(20 * time.Second)
	defer tick.Stop()
	if goos == "windows" {
		stopProgram()
		endCh := make(chan bool, 1)
		go func() {
			time.Sleep(1 * time.Second)
			startDataKit()
			endCh <- true
		}()
		for {
			select {
			case <-tick.C:
				l.Info("timeout")
				return
			case <-endCh:
				return
			}
		}
	}

	if goos == "darwin" {
		termSignal := make(chan os.Signal, 1)
		signal.Notify(termSignal)
		stopProgram()
		for {
			select {
			case s := <-termSignal:
				l.Infof("receive signal: %s", s)
				startDataKit()
				return
			case <-tick.C:
				l.Info("timeout")
				return
			}
		}
	}
}

func getBinName() string {
	bin := "datakit"
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	return bin
}

func runCmd(bin string, args ...string) (string, error) {
	cmd := exec.Command(bin, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func startDataKit() {
	l.Info("startDatakit")
	maxCount := 20
	for {
		time.Sleep(1 * time.Second)
		status, err := runCmd(program, "--status")

		if err == nil {
			l.Infof("current status: %s", status)
			if isStopped(status) {
				res, err := runCmd(program, "--start")
				if err != nil {
					l.Error(err)
				} else {
					l.Info(res)
					return
				}
			}
		} else {
			l.Error(err)
			l.Info("try to restart datakit ...") // for linux
			if _, err = runCmd(program, "--restart"); err != nil {
				l.Error(err)
			}
		}
		if maxCount < 0 {
			break
		}
		maxCount--
	}

	l.Warn("timeout to stop datakit, try to restart datakit")
	res, err := runCmd(program, "--restart")
	if err != nil {
		l.Error(err)
		return
	}
	l.Info(res)
}

func isStopped(status string) bool {
	matched, err := regexp.MatchString("(?i)stopped", status)
	if err != nil {
		return false
	}
	return matched
}

func stopProgram() {
	_, err := runCmd(program, "--stop")
	l.Info("stop datakit...")

	if err != nil { // ignore this error, parent process
		l.Warn("ignore err: ", err)
	}
}
