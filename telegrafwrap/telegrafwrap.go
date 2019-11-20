package telegrafwrap

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
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

	if err := startAgent(); err != nil {
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
			}
		case <-ctx.Done():
			stopAgent()
			return context.Canceled
		}
	}
}

func stopAgent() error {

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
			if err = prs.Kill(); err != nil {
				return err
			}
			time.Sleep(time.Millisecond * 100)
		}
	}

	return nil
}

func startAgent() error {

	stopAgent()

	env := os.Environ()
	procAttr := &os.ProcAttr{
		Env: env,
		Files: []*os.File{
			os.Stdin,
			os.Stdout,
			os.Stderr,
		},
	}

	p, err := os.StartProcess(agentPath(), []string{"agent", "-config", agentConfPath(true)}, procAttr)
	if err != nil {
		return err
	}

	ioutil.WriteFile(agentPidPath(), []byte(fmt.Sprintf("%d", p.Pid)), 0664)

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

func agentPath() string {
	return filepath.Join(config.ExecutableDir, "agent")
}
