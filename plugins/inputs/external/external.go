// Package external wraps all external command to collect various metrics
package external

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	configSample = `
[[inputs.external]]

	# 外部采集器名称
	name = 'some-external-inputs'  # required

	# 是否以后台方式运行外部采集器
	daemon = false

	# 如果以非 daemon 方式运行外部采集器，则以该间隔多次运行外部采集器
	#interval = '10s'

	# 运行外部采集器所需的环境变量
	#envs = ['LD_LIBRARY_PATH=/path/to/lib:$LD_LIBRARY_PATH',]

	# 外部采集器可执行程序路径(尽可能写绝对路径)
	cmd = "python" # required

	args = []
	`
)

var (
	inputName = "external"
	l         = logger.DefaultSLogger(inputName)
)

type ExernalInput struct {
	Name     string            `toml:"name"`
	Daemon   bool              `toml:"daemon"`
	Interval string            `toml:"interval"`
	Envs     []string          `toml:"envs"`
	Cmd      string            `toml:"cmd"`
	Args     []string          `toml:"args"`
	Tags     map[string]string `toml:"tags"`

	cmd      *exec.Cmd     `toml:"-"`
	duration time.Duration `toml:"-"`

	semStop          *cliutils.Sem // start stop signal
	semStopCompleted *cliutils.Sem // stop completed signal
}

func (*ExernalInput) Catalog() string {
	return "external"
}

func (*ExernalInput) SampleConfig() string {
	return configSample
}

func (ex *ExernalInput) precheck() error {
	ex.duration = time.Second * 10
	if ex.Interval != "" {
		du, err := time.ParseDuration(ex.Interval)
		if err != nil {
			l.Errorf("parse external input %s interval failed: %s", ex.Name, err.Error())
			return err
		}

		ex.duration = du
	}

	// TODO: check ex.Cmd is ok

	return nil
}

func (ex *ExernalInput) start() error {
	ex.cmd = exec.Command(ex.Cmd, ex.Args...) //nolint:gosec
	if ex.Envs != nil {
		ex.cmd.Env = ex.Envs
	}

	l.Debugf("starting cmd %s, envs: %+#v", ex.cmd.String(), ex.cmd.Env)
	if err := ex.cmd.Start(); err != nil {
		l.Errorf("start external input %s failed: %s", ex.Name, err.Error())
		return err
	}

	return nil
}

func (ex *ExernalInput) Run() {
	l = logger.SLogger(inputName)

	l.Infof("starting external input %s...", ex.Name)

	tagsStr := ""
	arr := []string{}
	for tagKey, tagVal := range ex.Tags {
		arr = append(arr, fmt.Sprintf("%s=%s", tagKey, tagVal))
	}
	if len(arr) > 0 {
		tagsStr = strings.Join(arr, ";")
	}

	if tagsStr != "" {
		ex.Args = append(ex.Args, []string{"--tags", tagsStr}...)
	}

	for {
		if err := ex.precheck(); err != nil {
			time.Sleep(time.Second)
			continue
		}
		break
	}

	if ex.Daemon {
		for {
			if err := ex.start(); err != nil { // start failed, retry
				time.Sleep(time.Second)
				continue
			}
			break
		}

		if err := datakit.MonitProc(ex.cmd.Process, ex.Name); err != nil { // blocking here...
			l.Errorf("datakit.MonitProc: %s", err.Error())
		}
		return
	}

	// non-daemon
	tick := time.NewTicker(ex.duration)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			_ = ex.start() //nolint:errcheck
		case <-datakit.Exit.Wait():
			l.Infof("external input %s exiting", ex.Name)
			return

		case <-ex.semStop.Wait():
			l.Infof("external input %s return", ex.Name)

			if ex.semStopCompleted != nil {
				ex.semStopCompleted.Close()
			}
			return
		}
	}
}

func (ex *ExernalInput) Terminate() {
	if ex.semStop != nil {
		ex.semStop.Close()

		// wait stop completed
		if ex.semStopCompleted != nil {
			for range ex.semStopCompleted.Wait() {
				return
			}
		}
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &ExernalInput{
			semStop:          cliutils.NewSem(),
			semStopCompleted: cliutils.NewSem(),
		}
	})
}
