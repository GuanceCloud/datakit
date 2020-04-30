package telegrafwrap

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
)

type (
	TelegrafSvr struct {
		Cfg *config.Config
		Pid int
	}
)

const agentSubDir = "embed"

var Svr = &TelegrafSvr{}

func (s *TelegrafSvr) Start() error {

	telegrafCfg, err := config.GenerateTelegrafConfig(s.Cfg)
	if err == config.ErrConfigNotFound {
		log.Printf("no sub service configuration found")
		return nil
	}

	if err != nil {
		return fmt.Errorf("fail to generate sub service config, %s", err)
	}

	agentCfgFile := s.agentConfPath(false)

	if err = ioutil.WriteFile(agentCfgFile, []byte(telegrafCfg), 0664); err != nil {
		return fmt.Errorf("fail to create file, %s", err.Error())
	}

	log.Printf("starting sub service...")

	if err := s.startAgent(); err != nil {
		return err
	}

	return nil
}

func (s *TelegrafSvr) StopAgent() error {

	pidpath := agentPidPath()

	log.Printf("[debug] read telegraf PID from %s", pidpath)
	piddata, err := ioutil.ReadFile(pidpath)
	if err != nil {
		return err
	}
	if string(piddata) == "" {
		return nil
	}
	npid, err := strconv.Atoi(string(piddata))
	if err != nil {
		return err
	}

	log.Printf("[debug] telegraf PID: %d", npid)
	if npid > 0 {
		return s.killProcessByPID(npid)
	}

	return nil
}

func (s *TelegrafSvr) startAgent(ctx context.Context) error {

	args := []string{}
	envs := []string{}

	if runtime.GOOS == "windows" {
		envs = []string{fmt.Sprintf(`TELEGRAF_CONFIG_PATH=%s`, s.agentConfPath(false))}
		args = []string{"-console"}
	}

	cmd := exec.Command(agentPath(), args...)
	cmd.Env = envs
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	s.Pid = cmd.Process.Pid

	return nil
}

func (s *TelegrafSvr) killProcessByPID(npid int) error {

	if CheckPid(npid) == nil {
		prs, err := os.FindProcess(npid)
		if err == nil && prs != nil {

			if err = prs.Kill(); err != nil {
				log.Printf("[error] kill telegraf failed: %s", err.Error())
				return err
			}
			time.Sleep(time.Millisecond * 20)
		}
	}

	return nil
}

func (s *TelegrafSvr) agentConfPath(quote bool) string {
	os.MkdirAll(filepath.Join(config.ExecutableDir, agentSubDir), 0775)
	path := filepath.Join(config.ExecutableDir, agentSubDir, "agent.conf")

	if quote {
		return fmt.Sprintf(`"%s"`, path)
	}
	return path
}

func agentPidPath() string {
	return filepath.Join(config.ExecutableDir, agentSubDir, "agent.pid")
}

func agentPath(win bool) string {

	if runtime.GOOS == "windows" {
		return filepath.Join(config.ExecutableDir, agentSubDir, "agent.exe")
	} else {
		return filepath.Join(config.ExecutableDir, agentSubDir, "agent")
	}
}
