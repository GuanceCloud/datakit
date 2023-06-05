// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package tdengine is input for tdengine database
package tdengine

import (
	"context"
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

var (
	_ inputs.InputV2       = &Input{}
	_ inputs.Input         = &Input{}
	_ inputs.ElectionInput = &Input{}
	_ inputs.PipelineInput = &Input{}
)

var (
	l         = logger.DefaultSLogger("tdengine")
	inputName = "tdengine"
)

const (
	sampleConfig = `
[[inputs.tdengine]]
  ## adapter restApi Addr, example: http://taosadapter.test.com  (Required)
  adapter_endpoint = "http://<FQND>:6041"
  user = "<userName>"
  password = "<pw>"

  ## log_files: TdEngine log file path or dirName (optional).
  ## log_files = ["tdengine_log_path.log"]
  ## pipeline = "tdengine.p"

  ## Set true to enable election
  election = true

  ## add tag (optional)
  [inputs.tdengine.tags]
  ## Different clusters can be distinguished by tag. Such as testing,product,local ,default is 'testing'
  # cluster_name = "testing"
  # some_tag = "some_value"
  # more_tag = "some_other_value"`

	//nolint:lll
	pipelineCfg = `
grok(_, '%{GREEDYDATA:temp}%{SPACE}TAOS_%{NOTSPACE:module}%{SPACE}%{NOTSPACE:level}%{SPACE}%{GREEDYDATA:http_url}')

if level != 'error' {
  grok(http_url,'"\\|%{SPACE}%{NOTSPACE:code}%{SPACE}\\|%{SPACE}%{NOTSPACE:cost_time}%{SPACE}\\|%{SPACE}%{NOTSPACE:client_ip}%{SPACE}\\|%{SPACE}%{NOTSPACE:method}%{SPACE}\\|%{SPACE}%{NOTSPACE:uri}%{SPACE}"')
  parse_duration(cost_time)

  duration_precision(cost_time, "ns", "ms")

}
group_in(level, ["error", "panic", "dpanic", "fatal","err","fat"], "error", status)
group_in(level, ["info", "debug", "inf", "bug"], "info", status)
group_in(level, ["warn", "war"], "warning", status)

`
)

type Input struct {
	AdapterEndpoint string            `toml:"adapter_endpoint"` // Adapter 访问地址，多个用逗号隔开
	User            string            `toml:"user"`
	Password        string            `toml:"password"`
	Tags            map[string]string `toml:"tags"`
	LogFiles        []string          `toml:"log_files"`
	Pipeline        string            `toml:"pipeline"`

	tdengine  *tdEngine
	inputName string
	tail      *tailer.Tailer

	Election bool `toml:"election"`
	pauseCh  chan bool
	semStop  *cliutils.Sem // start stop signal
}

func (i *Input) ElectionEnabled() bool {
	return i.Election
}

func (i *Input) Catalog() string {
	return "db"
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
}

func (i *Input) SampleConfig() string {
	return sampleConfig
}

func (i *Input) Terminate() {
	if i.semStop != nil {
		i.semStop.Close()
	}
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&Measurement{}}
}

// PipelineConfig : implement interface PipelineInput.
func (i *Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		inputName: pipelineCfg,
	}

	return pipelineMap
}

func (i *Input) RunPipeline() {
	if len(i.LogFiles) == 0 {
		return
	}

	opt := &tailer.Option{
		Source:     inputName,
		Service:    inputName,
		Pipeline:   i.Pipeline,
		GlobalTags: i.Tags,
		Done:       i.semStop.Wait(),
	}

	var err error
	i.tail, err = tailer.NewTailer(i.LogFiles, opt)
	if err != nil {
		l.Errorf("new tailer error: %v", err)
		return
	}
	l.Info("tailer start, logFile=%+v", i.LogFiles)
	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_tdengine"})
	g.Go(func(ctx context.Context) error {
		i.tail.Start()
		return nil
	})
}

//nolint:lll
func (i *Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		inputName: {
			"tdengine log 204": `08/22 13:44:34.290731 01081508 TAOS_ADAPTER info "| 204 |    1.641678ms |     172.16.5.29 | POST | /influxdb/v1/write?db=biz_ufssescxoxnuyzeypouezr_7d " model=web sessionID=48847`,
			"tdengine log 200": `08/22 13:44:41.108850 01081508 TAOS_ADAPTER info "| 200 |     2.45708ms |     172.16.5.29 | POST | /rest/sqlutc " model=web sessionID=48848`,
		},
	}
}

func (i *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:  inputName,
			Service: inputName,
			Pipeline: func() string {
				return i.Pipeline
			}(),
		},
	}
}

func (i *Input) exit() {
	i.tdengine.Stop()
	if i.tail != nil {
		// stop log 采集
		l.Info("stop tailer")
		i.tail.Close()
	}
}

func (i *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case i.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (i *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case i.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)

	i.tdengine = newTDEngine(i.User, i.Password, i.AdapterEndpoint, i.Election)

	globalTags = i.Tags

	// 1 checkHealth: show databases
	if !i.tdengine.CheckHealth(checkHealthSQL) {
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_tdengine"})
	g.Go(func(ctx context.Context) error {
		i.tdengine.run()
		return nil
	})
	l.Infof("TDEngine input started")

	for {
		select {
		case <-datakit.Exit.Wait():
			i.exit()
			l.Infof("%s exit", i.inputName)
			return
		case <-i.semStop.Wait():
			i.exit()
			l.Infof("%s return", i.inputName)
			return
		case f := <-i.pauseCh:
			i.tdengine.upstream = !f
		}
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Tags:      map[string]string{},
			inputName: inputName,
			Election:  true,
			pauseCh:   make(chan bool, inputs.ElectionPauseChannelLength),
			semStop:   cliutils.NewSem(),
		}
	})
}
