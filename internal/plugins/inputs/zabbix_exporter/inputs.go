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
  # [inputs.zabbix_exporter.tags]
    # key1 = "value1"
    # key2 = "value2"
    # ...

  # Zabbix server version 5.x.
  [inputs.zabbix_exporter.export_v5]
    # zabbix realTime exportDir path
    export_dir = "/data/zbx/datakit/"

`
)

var log = logger.DefaultSLogger(inputName)

type Input struct {
	Tags map[string]string `toml:"tags"`
	Ex   *ExporterV5       `toml:"export_v5"`

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
	if ipt.Ex != nil {
		err := ipt.Ex.InitExporter(ipt.feeder, ipt.Tags)
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
	return nil
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
