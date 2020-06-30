package telegrafwrap

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"go.uber.org/zap"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
)

type TelegrafSvr struct {
	Cfg *config.Config
}

var (
	Svr          = &TelegrafSvr{}
	l            *zap.SugaredLogger
	telegrafConf string
)

func (s *TelegrafSvr) Start() error {

	l = logger.SLogger("telegrafwrap")

	telegrafConf = filepath.Join(config.TelegrafDir, "agent.conf")

	conf, err := config.GenerateTelegrafConfig(s.Cfg)
	switch err {
	case nil:
	case config.ErrConfigNotFound:
		l.Info("no need to start sub service")
		return nil
	default:
		return fmt.Errorf("fail to generate sub service config, %s", err)
	}

	if err = ioutil.WriteFile(telegrafConf, []byte(conf), 0664); err != nil {
		return fmt.Errorf("fail to create file, %s", err.Error())
	}

	l.Info("starting telegraf...")

	proc, err := s.startAgent()
	if err != nil {
		l.Error(err)
		return err
	}

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
				// on windows, if os.FindProcess() ok, means the process is running
				l.Debugf("telegraf on PID %d ok", proc.Pid)
			default:
				if err := p.Signal(syscall.Signal(0)); err != nil {
					l.Errorf("signal 0 to telegraf failed: %s", err)
				}
			}

		case <-config.Exit.Wait():
			l.Info("exit, killing telegraf...")
			if err := proc.Kill(); err != nil { // XXX: should we wait here?
				l.Warnf("killing telegraf failed: %s, ignored", err)
			}

			l.Infof("killing telegraf (PID: %d) ok", proc.Pid)
			return nil
		}
	}

	return nil
}

func (s *TelegrafSvr) startAgent() (*os.Process, error) {

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
	fpath := filepath.Join(config.TelegrafDir, runtime.GOOS+"-"+runtime.GOARCH, "agent")
	if runtime.GOOS == "windows" {
		fpath = fpath + ".exe"
	}

	return fpath
}
