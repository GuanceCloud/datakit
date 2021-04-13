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
	l = logger.SLogger("telegraf_inputs")

	telegrafConf = filepath.Join(datakit.TelegrafDir, "agent.conf")

	l.Info("starting telegraf...")

	proc, err := doStart()
	if err != nil {
		return err
	}

	return datakit.MonitProc(proc, "telegraf")
}

func doStart() (*os.Process, error) {
	var p = &os.Process{}
	telegrafBin := agentPath()

	if runtime.GOOS == datakit.OSWindows {
		env := os.Environ()
		env = append(env, fmt.Sprintf(`TELEGRAF_CONFIG_PATH=%s`, telegrafConf))

		cmd := exec.Command(telegrafBin, "-console")
		cmd.Env = env

		// XXX: under windows, we must redirect cmd stdout/stderr to os, or
		// the restarting of telegraf will timeout.

		// XXX: this makes me hard to get the telegraf startup error message(i.e., config error)

		// TODO: we should check all telegraf config before starting it
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Start()
		if err != nil {
			//l.Error("start telegraf failed: %s, %s", err.Error(), string(out))
			l.Error("start telegraf failed: %s", err.Error())
			return nil, err
		}

		p = cmd.Process
	} else {
		cmd := exec.Command(telegrafBin, "-config", telegrafConf)
		go func() {
			var err error
			out, err := cmd.CombinedOutput()
			if err != nil {
				l.Warnf("%s, %s", err.Error(), string(out))
			}
		}()
		time.Sleep(time.Second)
		p = cmd.Process
	}

	l.Infof("telegraf PID: %+#v", p)
	//time.Sleep(time.Second)
	return p, nil
}

func agentPath() string {
	fpath := filepath.Join(datakit.TelegrafDir, runtime.GOOS+"-"+runtime.GOARCH, "agent")
	if runtime.GOOS == datakit.OSWindows {
		fpath += ".exe"
	}

	return fpath
}
