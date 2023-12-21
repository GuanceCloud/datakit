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
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
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
	feeder   dkio.Feeder
	Tagger   datakit.GlobalTagger
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (*Input) Catalog() string {
	return "db"
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&Measurement{}}
}

// PipelineConfig : implement interface PipelineInput.
func (ipt *Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		inputName: pipelineCfg,
	}

	return pipelineMap
}

func (ipt *Input) RunPipeline() {
	if len(ipt.LogFiles) == 0 {
		return
	}

	opt := &tailer.Option{
		Source:     inputName,
		Service:    inputName,
		Pipeline:   ipt.Pipeline,
		GlobalTags: inputs.MergeTags(ipt.Tagger.HostTags(), ipt.Tags, ""),
		Done:       ipt.semStop.Wait(),
	}

	var err error
	ipt.tail, err = tailer.NewTailer(ipt.LogFiles, opt)
	if err != nil {
		l.Errorf("new tailer error: %v", err)
		return
	}
	l.Info("tailer start, logFile=%+v", ipt.LogFiles)
	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_tdengine"})
	g.Go(func(ctx context.Context) error {
		ipt.tail.Start()
		return nil
	})
}

//nolint:lll
func (ipt *Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		inputName: {
			"tdengine log 204": `08/22 13:44:34.290731 01081508 TAOS_ADAPTER info "| 204 |    1.641678ms |     172.16.5.29 | POST | /influxdb/v1/write?db=biz_ufssescxoxnuyzeypouezr_7d " model=web sessionID=48847`,
			"tdengine log 200": `08/22 13:44:41.108850 01081508 TAOS_ADAPTER info "| 200 |     2.45708ms |     172.16.5.29 | POST | /rest/sqlutc " model=web sessionID=48848`,
		},
	}
}

func (ipt *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:  inputName,
			Service: inputName,
			Pipeline: func() string {
				return ipt.Pipeline
			}(),
		},
	}
}

func (ipt *Input) exit() {
	ipt.tdengine.Stop()
	if ipt.tail != nil {
		// stop log 采集
		l.Info("stop tailer")
		ipt.tail.Close()
	}
}

func (ipt *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (ipt *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	ipt.tdengine = newTDEngine(ipt)

	// 1 checkHealth: show databases
	if !ipt.tdengine.CheckHealth(checkHealthSQL) {
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_tdengine"})
	g.Go(func(ctx context.Context) error {
		ipt.tdengine.run()
		return nil
	})
	l.Infof("TDEngine input started")

	for {
		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Infof("%s exit", ipt.inputName)
			return
		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Infof("%s return", ipt.inputName)
			return
		case f := <-ipt.pauseCh:
			ipt.tdengine.upstream = !f
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
			feeder:    dkio.DefaultFeeder(),
			Tagger:    datakit.DefaultGlobalTagger(),
		}
	})
}
