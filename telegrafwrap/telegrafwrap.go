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

const (
	agentSubDir = "embed"
)

var (
	Svr = &TelegrafSvr{}

	l *zap.SugaredLogger

	agentPID int = -1
)

func (s *TelegrafSvr) Start() error {

	l = logger.SLogger("telegrafwrap")

	telegrafCfg, err := config.GenerateTelegrafConfig(s.Cfg)
	switch err {
	case config.ErrConfigNotFound:
		l.Info("no need to start sub service")
		return nil
	case nil:
	default:
		return fmt.Errorf("fail to generate sub service config, %s", err)
	}

	agentCfgFile := s.agentConfPath(false)

	if err = ioutil.WriteFile(agentCfgFile, []byte(telegrafCfg), 0664); err != nil {
		return fmt.Errorf("fail to create file, %s", err.Error())
	}

	l.Info("starting sub service...")

	proc, err := s.startAgent()
	if err != nil {
		l.Error(err)
		return err
	}

	tick := time.NewTicker(time.Second)

	for {
		select {
		case <-tick.C:
			p, err := os.FindProcess(proc.Pid)
			if err != nil {
				l.Error(err)
				continue
			}

			if err := p.Signal(syscall.Signal(0)); err != nil {
				l.Errorf("signal 0 to telegraf failed: %s", err)
			}

		case <-config.Exit.Wait():
			l.Info("exit, killing telegraf...")
			if err := proc.Kill(); err != nil { // XXX: should we wait here?
				l.Warnf("killing telegraf failed: %s, ignored", err)
			}

			return nil
		}
	}

	return nil
}

func (s *TelegrafSvr) startAgent() (*os.Process, error) {

	env := os.Environ()
	if runtime.GOOS == "windows" {
		env = append(env, fmt.Sprintf(`TELEGRAF_CONFIG_PATH=%s`, s.agentConfPath(false)))
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
		p, err = os.StartProcess(agentPath(), []string{"agent", "-config", s.agentConfPath(false)}, procAttr)
		if err != nil {
			return nil, err
		}
	}

	l.Infof("agent start on %d", p.Pid)
	time.Sleep(time.Millisecond * 20)
	return p, nil
}

func (s *TelegrafSvr) agentConfPath(quote bool) string {
	os.MkdirAll(filepath.Join(config.InstallDir, agentSubDir), 0775)
	path := filepath.Join(config.InstallDir, agentSubDir, "agent.conf")

	if quote {
		return fmt.Sprintf(`"%s"`, path)
	}
	return path
}

func agentPath() string {
	fpath := filepath.Join(config.InstallDir, agentSubDir, runtime.GOOS+"-"+runtime.GOARCH, "agent")
	if runtime.GOOS == "windows" {
		fpath = fpath + ".exe"
	}

	return fpath
}
