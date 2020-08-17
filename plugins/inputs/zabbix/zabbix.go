package zabbix

import (
	"os"
	"path/filepath"
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	zabbixConfigSample = `
#[[inputs.zabbix]]
#  active        = true
#  dbType        = "mysql"
#  dbAddress     = "zabbix:zabbix@tcp(127.0.0.1:3306)/zabbix"
#  startdate     = "2020-01-01T00:00:00"
#  hoursperbatch = 720
#  interval      = "60s"
#  registry      = "influxdb-zabbix"
#  [inputs.zabbix.tags]
#    tag1 = "tag1"
#    tag2 = "tag2"
#    tag3 = "tag3"
`
)

var (
	defaultDataDir   = "data"
	defaultInterval  = "60s"
	defaultZabbixDir = "zabbix"
	defaultStartDate = "2019-12-02T00:00:00"
	locker           sync.Mutex
	inputName        = "zabbix"
)

type Zabbix struct {
	Active        bool
	DbType        string
	DbAddress     string
	Startdate     string
	Hoursperbatch int
	Interval      interface{}
	Registry      string
	Tags          map[string]string
}

func (z *Zabbix) Catalog() string {
	return `zabbix`
}

func (z *Zabbix) SampleConfig() string {
	return zabbixConfigSample
}

func (z *Zabbix) Run() {
	if !z.Active {
		return
	}

	regPath := filepath.Join(datakit.InstallDir, defaultDataDir, defaultZabbixDir, z.Registry)
	z.Registry = regPath

	if z.DbType == "mysql" {
		z.DbAddress += "?sql_mode='PIPES_AS_CONCAT'"
	}

	if z.Interval == nil {
		z.Interval = defaultInterval
	}

	input := ZabbixInput{*z}
	output := ZabbixOutput{io.NamedFeed}
	p := &ZabbixParam{input, output, logger.SLogger("zabbix")}
	p.log.Info("yarn zabbix started...")
	p.mkZabbixDataDir()
	p.gather()
}

func (p *ZabbixParam) mkZabbixDataDir() {
	dataDir := filepath.Join(datakit.InstallDir, defaultDataDir)
	zabbixDir := filepath.Join(dataDir, defaultZabbixDir)

	if !PathExists(dataDir) {
		return
	}
	if PathExists(zabbixDir) {
		return
	}

	locker.Lock()
	defer locker.Unlock()
	if PathExists(zabbixDir) {
		return
	}

	err := os.MkdirAll(zabbixDir, 0666)
	if err != nil {
		p.log.Error("Mkdir zabbix err: %s", err.Error())
	}
}
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		z := &Zabbix{}
		return z
	})
}
