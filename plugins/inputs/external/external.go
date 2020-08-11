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

	# 如果以非 daemon 方式运行外部采集器，则以该间隔多次运行外部采集器。否则该配置无效
	#interval = '10s'

	# 运行外部采集器所需的环境变量
	#envs = ['LD_LIBRARY_PATH=/path/to/lib:$LD_LIBRARY_PATH',]

	# 外部采集器运行命令（任何命令均可，不可使用组合命令，如 'ps -ef && echo ok | print'）
	cmd = "python your-python-script.py -cfg your-config.conf" # required

	# 本采集器不支持自定义 tag，所有自定义 tag 追加应该在外部采集器中自行追加
	`
)

var (
	inputName = "external"
	l         = logger.DefaultSLogger(inputName)
)

type externalInput struct {
	Name     string   `toml:"name"`
	Daemon   bool     `toml:"daemon"`
	Interval string   `toml:"interval"`
	Envs     []string `toml:"envs"`
	Cmd      string   `toml:"cmd"`

	cmd      *exec.Cmd     `toml:"-"`
	bin      string        `toml:"-"`
	args     []string      `toml:"-"`
	duration time.Duration `toml:"-"`
}

func (_ *externalInput) Catalog() string {
	return "external"
}

func (_ *externalInput) SampleConfig() string {
	return configSample
}

func (ex *externalInput) precheck() error {
	elems := strings.Split(ex.Cmd, " ")
	if len(elems) == 0 {
		l.Errorf("external input %s: empty Cmd", ex.Name)
		return fmt.Errorf("invalid cmd")
	}

	ex.bin = elems[0]

	if len(elems) > 1 {
		ex.args = elems[1:]
	}

	ex.duration = time.Second * 10
	if ex.Interval != "" {
		du, err := time.ParseDuration(ex.Interval)
		if err != nil {
			l.Errorf("parse external input %s interval failed: %s", ex.Name, err.Error())
			return err
		}

		ex.duration = du
	}

	return nil
}

func (ex *externalInput) start() error {
	ex.cmd = exec.Command(ex.bin, ex.args...)
	ex.cmd.Env = ex.Envs

	if err := ex.cmd.Start(); err != nil {
		l.Errorf("start external input %s failed: %s", ex.Name, err.Error())
		return err
	}

	return nil
}

func (ex *externalInput) Run() {
	l = logger.SLogger(inputName)

	l.Infof("starting external input %s...", ex.Name)

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
		return &externalInput{}
	})
}
