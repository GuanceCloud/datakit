// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package exporter collect RealTime data.
package exporter

import (
	"context"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var _ inputs.InputV2 = &Input{}

const (
	inputName    = "zabbix_exporter"
	sampleConfig = `
[[inputs.zabbix_exporter]]
  ## zabbix server web.
  localhostAddr = "http://localhost/zabbix/api_jsonrpc.php"
  user_name = "Admin"
  user_pw = "zabbix"
  
  ## measurement yaml Dir
  measurement_config_dir = "/data/zbx/yaml"

  ## exporting object.default is item. all is <trigger,item,trends>. 
  objects = "item"

  ## update items and interface data.
  ## like this: All data is updated at 2 o'clock every day.
  crontab = "0 2 * * *"

  # [inputs.zabbix_exporter.tags]
    # key1 = "value1"
    # key2 = "value2"
    # ...

  ## mysql database:zabbix , tables: items,interface.
  #[inputs.zabbix_exporter.mysql]
  #  db_host = "192.168.10.12"
  #  db_port = "3306"
  #  user = "root"
  #  pw = "123456"

  # Zabbix server version 4.x - 7.x
  [inputs.zabbix_exporter.export_v5]
    # zabbix realTime exportDir path
    export_dir = "/data/zbx/datakit/"
    # 4.0~4.9 is v4
    # 5.0~7.x is v5
    module_version = "v5"
`
)

var log = logger.DefaultSLogger(inputName)

type Input struct {
	LocalhostAddr        string            `toml:"localhostAddr"`
	UserName             string            `toml:"user_name"`
	UserPW               string            `toml:"user_pw"`
	MeasurementConfigDir string            `toml:"measurement_config_dir"`
	Tags                 map[string]string `toml:"tags"`
	Ex                   *ExporterV5       `toml:"export_v5"`
	Mysql                *Mysql            `toml:"mysql"`
	Objects              string            `toml:"objects"`
	Crontab              string            `toml:"crontab"`

	feeder  dkio.Feeder
	semStop *cliutils.Sem
}

func (ipt *Input) Catalog() string {
	return inputName
}

func (ipt *Input) Run() {
	log = logger.SLogger(inputName)
	if ipt.feeder == nil {
		ipt.feeder = dkio.DefaultFeeder()
	}
	// 先初始化缓存对象。在初始化 exporterV5.
	cd := &CacheData{
		api: &ZabbixAPI{
			server: ipt.LocalhostAddr,
			user:   ipt.UserName,
			pw:     ipt.UserPW,
		},
		Items:        make(map[int64]*ItemC),
		Interfaces:   make(map[int64]*InterfaceC),
		Measurements: make(map[string]*Measurement),
	}

	err := cd.start(ipt.Mysql, ipt.MeasurementConfigDir, ipt.Crontab)
	if err != nil {
		log.Errorf("start zabbix_exporter err =%v", err)
		return
	}

	if ipt.Ex != nil {
		err := ipt.Ex.InitExporter(ipt.feeder, ipt.Tags, cd, ipt.Objects)
		if err != nil {
			log.Errorf("init exporter err = %v", err)
		} else {
			g := goroutine.NewGroup(goroutine.Option{Name: inputName})
			g.Go(func(ctx context.Context) error {
				ipt.Ex.collect()
				return nil
			})
		}
	}

	select {
	case <-datakit.Exit.Wait():
		ipt.Terminate()
		log.Infof("%s exit", inputName)
		return
	case <-ipt.semStop.Wait():
		ipt.Terminate()
		log.Infof("%s exit", inputName)
		return
	}
}

func (ipt *Input) SampleConfig() string {
	return sampleConfig
}

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{}
}

func (ipt *Input) AvailableArchs() []string {
	return []string{datakit.OSLabelLinux}
}

func (ipt *Input) Terminate() {
	if ipt.Ex != nil {
		close(ipt.Ex.stopChan)
	}
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			semStop: cliutils.NewSem(),
			feeder:  dkio.DefaultFeeder(),
			Tags:    make(map[string]string),
		}
	})
}
