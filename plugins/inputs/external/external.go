package external

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

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
}

func (_ *ExernalInput) Catalog() string {
	return "external"
}

func (_ *ExernalInput) SampleConfig() string {
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
	ex.cmd = exec.Command(ex.Cmd, ex.Args...)
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
			if err := ex.start(); err != nil {
				time.Sleep(time.Second)
			}
			break
		}

		datakit.MonitProc(ex.cmd.Process, ex.Name) // blocking here...
		return
	}

	// non-daemon
	tick := time.NewTicker(ex.duration)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			ex.start()
		case <-datakit.Exit.Wait():
			l.Infof("external input %s exiting", ex.Name)
			return
		}
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &ExernalInput{}
	})
}
