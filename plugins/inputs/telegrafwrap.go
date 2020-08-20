package inputs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	telegrafConf string
)

func startTelegraf() {

	telegrafConf = filepath.Join(datakit.TelegrafDir, "agent.conf")

	l.Info("starting telegraf...")

	proc, err := doStart()
	if err != nil {
		l.Error(err)
		return
	}

	datakit.MonitProc(proc, "telegraf")
}

func doStart() (*os.Process, error) {

	env := os.Environ()
	if runtime.GOOS == "windows" {
		env = append(env, fmt.Sprintf(`TELEGRAF_CONFIG_PATH=%s`, telegrafConf))
	}
	procAttr := &os.ProcAttr{
		Env: env,
		Files: []*os.File{
			os.Stdin,
			os.Stdout,
			os.Stderr,
		},
	}

	var p *os.Process
	var err error

	if runtime.GOOS == "windows" {

		cmd := exec.Command(agentPath(), "-console")
		cmd.Env = env
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			return nil, err
		}
		p = cmd.Process

	} else {
		p, err = os.StartProcess(agentPath(), []string{"agent", "-config", telegrafConf}, procAttr)
		if err != nil {
			return nil, err
		}
	}

	l.Infof("telegraf PID: %d", p.Pid)
	time.Sleep(time.Millisecond * 20)
	return p, nil
}

func agentPath() string {
	fpath := filepath.Join(datakit.TelegrafDir, runtime.GOOS+"-"+runtime.GOARCH, "agent")
	if runtime.GOOS == "windows" {
		fpath = fpath + ".exe"
	}

	return fpath
}
