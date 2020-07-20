// +build linux

package oraclemonitor

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"go.uber.org/zap"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	configSample = `
#libPath = ''
#[[instances]]
#  ## 采集的频度，最小粒度5m
#  interval = '10s'
#  ## 指标集名称，默认值oracle_monitor
#  metricName = ''
#  ## 实例ID(非必要属性)
#  instanceId = ''
#  ## # 实例描述(非必要属性)
#  instanceDesc = ''
#  ## oracle实例地址(ip)
#  host = ''
#  ## oracle监听端口
#  port = ''
#  ## 帐号
#  username = ''
#  ## 密码
#  password = ''
#  ## oracle的服务名
#  server = ''
#  ## 实例类型 例如 单实例、DG、RAC 等，非必要属性
#  type= 'singleInstance'
#`
)

var (
	l *zap.SugaredLogger
)

type Instance struct {
	Metric   string `json:"metricName" toml:"metricName"`
	Interval string `json:"interval" toml:"interval"`

	InstanceId string `json:"instanceId" toml:"instanceId"`
	User       string `json:"username" toml:"username"`
	Password   string `json:"password" toml:"password"`
	Desc       string `json:"instanceDesc" toml:"instanceDesc"`
	Host       string `json:"host" toml:"host"`
	Port       string `json:"port" toml:"port"`
	Server     string `json:"server" toml:"server"`
	Type       string `json:"type" toml:"type"`

	DB               *sql.DB       `json:"-"`
	IntervalDuration time.Duration `json:"-"`
}

type oraclemonitor struct {
	LibPath   string      `toml:"libPath"`
	Instances []*Instance `toml:"instances"`
}

func (_ *oraclemonitor) Catalog() string {
	return "oracle"
}

func (_ *oraclemonitor) SampleConfig() string {
	return configSample
}

func (o *oraclemonitor) Run() {
	l = logger.SLogger("oraclemonitor")
	l.Info("oraclemonitor started")

	l.Info("starting external oraclemonitor...")

	bin := filepath.Join(datakit.InstallDir, "external", "oraclemonitor")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	if _, err := os.Stat(bin); err != nil {
		l.Error("check %s failed: %s", bin, err.Error())
		return
	}

	args := []string{
		"-cfg", filepath.Join(datakit.ConfdDir, o.Catalog(), "oraclemonintor.conf"),
		"-log-level", config.Cfg.MainCfg.LogLevel,
	}

	cmd := exec.Command(bin, args...)
	cmd.Env = []string{ // we need oracle instantclient_xx_xx lib
		fmt.Sprintf("LD_LIBRARY_PATH=%s:$LD_LIBRARY_PATH", o.LibPath),
	}

	l.Infof("starting process %+#v", cmd)
	if err := cmd.Start(); err != nil {
		l.Errorf("starting oraclemonintor failed: %s", err.Error())
		return
	}

	l.Infof("oraclemonintor PID: %d", cmd.Process.Pid)
	datakit.MonitProc(cmd.Process, "oraclemonintor") // blocking
}

func init() {
	inputs.Add("oraclemonintor", func() inputs.Input {
		return &oraclemonitor{}
	})
}
