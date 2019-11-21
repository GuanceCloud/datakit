package telegrafwrap

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/log"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pid"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/uploader"
)

func init() {

	service.Add("agent", func(logger log.Logger) service.Service {

		telcfg, err := config.GenerateTelegrafConfig()
		if err != nil {
			if err != config.ErrNoTelegrafConf {
				logger.Errorf("%s", err.Error())
			}
			return nil
		}

		if err = ioutil.WriteFile(agentConfPath(false), []byte(telcfg), 0664); err != nil {
			logger.Errorf("%s", err.Error())
			return nil
		}

		return &TelegrafSvr{
			logger: logger,
		}
	})
}

var (
	errorAgentBeKilled = errors.New("agent has been killed")
)

type (
	TelegrafSvr struct {
		logger log.Logger
	}
)

func (s *TelegrafSvr) Start(ctx context.Context, up uploader.IUploader) error {

	s.logger.Info("Starting agent...")
	defer func() {
		s.logger.Info("agent done")
	}()

	if err := startAgent(ctx, s.logger); err != nil {
		s.logger.Errorf("start agent fail: %s", err.Error())
		return err
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			piddata, err := ioutil.ReadFile(agentPidPath())
			if err != nil {
				return errorAgentBeKilled
			}

			npid, err := strconv.Atoi(string(piddata))
			if err != nil || npid <= 2 {
				return errorAgentBeKilled
			}

			if pid.CheckPid(npid) != nil {
				return errorAgentBeKilled
			} else {
				_, err := os.FindProcess(npid)
				if err != nil {
					s.logger.Warnf("agent has quited: %s", err.Error())
				}
			}
		case <-ctx.Done():
			s.logger.Info("start quit agent")
			stopAgent(s.logger)
			s.logger.Info("end quit agent")
			return context.Canceled
		}
	}
}

func stopAgent(l log.Logger) error {

	piddata, err := ioutil.ReadFile(agentPidPath())
	if err != nil {
		return err
	}
	npid, err := strconv.Atoi(string(piddata))
	if err != nil {
		return err
	}

	if pid.CheckPid(npid) == nil {
		prs, err := os.FindProcess(npid)
		if err == nil && prs != nil {
			l.Info("find agent by pid")
			if err = prs.Signal(syscall.SIGTERM); err != nil {
				l.Error("kill agent failed")
				return err
			}
			l.Info("killed agent by pid")

			time.Sleep(time.Millisecond * 100)
		}
	}

	return nil
}

func startAgent(ctx context.Context, l log.Logger) error {

	stopAgent(l)

	env := os.Environ()
	if runtime.GOOS == "windows" {
		env = append(env, fmt.Sprintf(`TELEGRAF_CONFIG_PATH=%s`, agentConfPath(false)))
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

		// pg := filepath.Join(config.ExecutableDir, "agent_debug.log")

		// f, _ := os.Create(pg)
		// procAttr.Files = []*os.File{
		// 	nil,
		// 	f,
		// 	f,
		// }

		// p, err = os.StartProcess(agentPath(true), []string{}, procAttr)
		// if err != nil {
		// 	return err
		// }

		cmd := exec.Command(agentPath(true), "-console")
		//cmd := exec.CommandContext(ctx, agentPath(true), "-console")
		cmd.Env = env
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			return err
		}
		p = cmd.Process

		l.Infof("agent is running, %v", p.Pid)

	} else {
		p, err = os.StartProcess(agentPath(false), []string{"agent", "-config", agentConfPath(false)}, procAttr)
		if err != nil {
			return err
		}
	}

	if p != nil {
		ioutil.WriteFile(agentPidPath(), []byte(fmt.Sprintf("%d", p.Pid)), 0664)
	}

	time.Sleep(time.Millisecond * 100)

	return nil
}

func agentConfPath(quote bool) string {
	if quote {
		return fmt.Sprintf(`"%s"`, filepath.Join(config.ExecutableDir, "agent.conf"))
	} else {
		return filepath.Join(config.ExecutableDir, "agent.conf")
	}
}

func agentPidPath() string {
	return filepath.Join(config.ExecutableDir, "agent.pid")
}

func agentPath(win bool) string {
	if win {
		return filepath.Join(config.ExecutableDir, "agent.exe")
	}
	return filepath.Join(config.ExecutableDir, "agent")
}
