package telegrafwrap

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/log"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/uploader"
)

func init() {

	service.Add("agent", func(logger log.Logger) service.Service {
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

	telcfg, err := config.GenerateTelegrafConfig()
	if err != nil {
		s.logger.Errorf("%s", err.Error())
		return err
	}

	if err = ioutil.WriteFile(agentConfPath(), []byte(telcfg), 0664); err != nil {
		s.logger.Errorf("%s", err.Error())
		return err
	}

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
			pid, err := ioutil.ReadFile(agentPidPath())
			if err != nil {
				return errorAgentBeKilled
			}

			npid, err := strconv.Atoi(string(pid))
			if err != nil || npid <= 2 {
				return errorAgentBeKilled
			}

			if checkPid(npid) != nil {
				return errorAgentBeKilled
			}
		case <-ctx.Done():
			stopAgent()
			return context.Canceled
		}
	}
}

func stopAgent() error {

	pid, err := ioutil.ReadFile(agentPidPath())
	if err != nil {
		return err
	}
	npid, err := strconv.Atoi(string(pid))
	if err != nil {
		return err
	}

	if checkPid(npid) == nil {
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

	p, err := os.StartProcess(agentPath(), []string{"agent", "--config", agentConfPath()}, procAttr)
	if err != nil {
		return err
	}

	ioutil.WriteFile(agentPidPath(), []byte(fmt.Sprintf("%d", p.Pid)), 0664)

	time.Sleep(time.Millisecond * 100)

	return nil
}

func checkPid(pid int) error {
	return syscall.Kill(pid, 0)
}

func agentConfPath() string {
	return filepath.Join(config.ExecutableDir, "agent.conf")
}

func agentPidPath() string {
	return filepath.Join(config.ExecutableDir, "agent.pid")
}

func agentPath() string {
	return filepath.Join(config.ExecutableDir, "agent")
}
