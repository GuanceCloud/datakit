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
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
)

type (
	TelegrafSvr struct {
		Cfg *config.Config
	}
)

const (
	agentSubDir = "embed"
)

var (
	Svr = &TelegrafSvr{}

	agentPID int = -1
)

func (s *TelegrafSvr) Start(ctx context.Context) error {

	telegrafCfg, err := config.GenerateTelegrafConfig(s.Cfg)
	if err == config.ErrConfigNotFound {
		log.Printf("no need to start sub service")
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

	if err := s.startAgent(ctx); err != nil {
		return err
	}

	//检查telegraf是否被kill掉或崩溃
	go func(ctx context.Context) {

		for {

			internal.SleepContext(ctx, 3*time.Second)

			select {
			case <-ctx.Done():
				return
			default:
			}

			ps, err := os.FindProcess(agentPID)
			if err != nil || ps == nil {
				log.Printf("W! sub service(%v) not found: %s", agentPID, err)
			}
		}
	}(ctx)

	return nil
}

func (s *TelegrafSvr) StopAgent() error {

	if agentPID <= 0 {
		return nil
	}

	log.Printf("stopping sub service %v...", agentPID)

	if err := KillProcess(agentPID); err != nil {
		log.Printf("E! fail to stop sub service, %s", err)
		return err
	} else {
		log.Printf("sub service stopped successfully")
	}

	return nil
}

func (s *TelegrafSvr) startAgent(ctx context.Context) error {

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
			return err
		}
		p = cmd.Process

	} else {
		p, err = os.StartProcess(agentPath(), []string{"agent", "-config", s.agentConfPath(false)}, procAttr)
		if err != nil {
			return err
		}
	}

	if p != nil {
		agentPID = p.Pid
		log.Printf("agent start on %d", agentPID)
	}

	time.Sleep(time.Millisecond * 20)

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

func agentPath() string {
	fpath := filepath.Join(config.ExecutableDir, agentSubDir, runtime.GOOS+"-"+runtime.GOARCH, "agent")
	if runtime.GOOS == "windows" {
		fpath = fpath + ".exe"
	}

	return fpath
}
