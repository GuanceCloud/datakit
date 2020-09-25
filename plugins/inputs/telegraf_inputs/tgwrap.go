package telegraf_inputs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	telegrafConf string
	l            = logger.DefaultSLogger("telegraf_inputs")
)

func StartTelegraf() error {

	telegrafConf = filepath.Join(datakit.TelegrafDir, "agent.conf")

	l.Info("starting telegraf...")

	proc, err := doStart()
	if err != nil {
		return err
	}

	return datakit.MonitProc(proc, "telegraf")
}

func doStart() (*os.Process, error) {

	env := os.Environ()
	if runtime.GOOS == datakit.OSWindows {
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
	telegrafBin := agentPath()

	if runtime.GOOS == datakit.OSWindows {

		cmd := exec.Command(telegrafBin, "-console")
		cmd.Env = env
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			return nil, err
		}

		p = cmd.Process
	} else {
		var err error
		p, err = os.StartProcess(telegrafBin, []string{"agent", "-config", telegrafConf}, procAttr)
		if err != nil {
			l.Errorf("start telegraf failed: %s", err.Error())
			return nil, err
		}
	}

	l.Infof("telegraf PID: %d", p.Pid)
	time.Sleep(time.Second)
	return p, nil
}

func agentPath() string {
	fpath := filepath.Join(datakit.TelegrafDir, runtime.GOOS+"-"+runtime.GOARCH, "agent")
	if runtime.GOOS == datakit.OSWindows {
		fpath += ".exe"
	}

	return fpath
}
