package zabbix

import (
	"log"
	"io"
	"context"
	"path/filepath"
	"reflect"

	"github.com/influxdata/telegraf"
	zlog "github.com/siddontang/go-log/log"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	pluginName   = "zabbix"
	registryPath string
	mapTables    = make(MapTable)
	zabbix       *Zabbix
	stopChan     chan bool
)


type zabbixLogWriter struct {
	io.Writer
}

type Zabix struct {
	Address string
}

type Table struct {
	Name               string
	Active             bool
	Interval           int
	Startdate          string
	Hoursperbatch      int
	Outputrowsperbatch int
}
type RegistryName struct {
	FileName string
}
//type Polling struct {
//	Interval        int `toml:"interval"`
//	IntervalIfError int `toml:"intervaliferror"`
//}

type Zabbix struct {
	//Polling  Polling           `toml:"polling"`
	Zabbix   map[string]*Zabix `toml:"zabbix"`
	Tables   map[string]*Table `toml:"tables"`
	Registry RegistryName

	ctx  context.Context
	cfun context.CancelFunc
	acc  telegraf.Accumulator
}

func (z *Zabbix) SampleConfig() string {
	return zabbixConfigSample
}

func (z *Zabbix) Description() string {
	return "Convert Zabbix Database to Dataway"
}

func (z *Zabbix) Gather(telegraf.Accumulator) error {
	return nil
}

func (z *Zabbix) globalInit() {
	zabbix = z
	stopChan = make(chan bool, len(z.Tables))
	registryPath = filepath.Join(config.ExecutableDir, "data", pluginName, z.Registry.FileName)

	ReadRegistry(registryPath, &mapTables)
}

func (z *Zabbix) Start(acc telegraf.Accumulator) error {
	z.globalInit()
	setupLogger()

	var provider string = (reflect.ValueOf(zabbix.Zabbix).MapKeys())[0].String()
	var address string = zabbix.Zabbix[provider].Address

	if provider == "mysql" {
		address += "?sql_mode='PIPES_AS_CONCAT'"
	}

	log.Printf("I! [Zabbix] start")

	for _, table := range zabbix.Tables {
		if table.Active {
			input := ZabbixInput{
				provider,
				address,
				table.Name,
				table.Interval,
				table.Hoursperbatch}

			output := ZabbixOutput{
				z.ctx,
				z.cfun,
				acc}

			p := &ZabbixParam{input, output}
			go p.gather()
		}
	}

	return nil
}

func (z *Zabbix) Stop() {
	for range z.Tables {
		stopChan <- true
	}
}

func setupLogger() {
	loghandler, _ := zlog.NewStreamHandler(&zabbixLogWriter{})
	zlogger := zlog.New(loghandler, 0)
	zlog.SetLevel(zlog.LevelDebug)
	zlog.SetDefaultLogger(zlogger)
}

func init() {
	inputs.Add(pluginName, func() telegraf.Input {
		z := &Zabbix{}
		z.ctx, z.cfun = context.WithCancel(context.Background())
		return z
	})
}
