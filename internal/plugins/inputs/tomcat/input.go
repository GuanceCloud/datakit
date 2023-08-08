// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package tomcat collect Tomcat metrics.
package tomcat

import (
	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

const (
	inputName = "tomcat"
)

var l = logger.DefaultSLogger(inputName)

type tomcatlog struct {
	Files             []string `toml:"files"`
	Pipeline          string   `toml:"pipeline"`
	IgnoreStatus      []string `toml:"ignore"`
	CharacterEncoding string   `toml:"character_encoding"`
	MultilineMatch    string   `toml:"multiline_match"`
}

type Input struct {
	inputs.JolokiaAgent
	Log  *tomcatlog        `toml:"log"`
	Tags map[string]string `toml:"tags"`
}

func (*Input) Catalog() string {
	return inputName
}

func (*Input) SampleConfig() string {
	return tomcatSampleCfg
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&TomcatGlobalRequestProcessorM{},
		&TomcatJspMonitorM{},
		&TomcatThreadPoolM{},
		&TomcatServletM{},
		&TomcatCacheM{},
		&TomcatM{},
	}
}

func (*Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		inputName: pipelineCfg,
	}
	return pipelineMap
}

//nolint:lll
func (i *Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		inputName: {
			"Tomcat access log":   `0:0:0:0:0:0:0:1 - admin [24/Feb/2015:15:57:10 +0530] "GET /manager/images/tomcat.gif HTTP/1.1" 200 2066`,
			"Tomcat Catalina log": `06-Sep-2021 22:33:30.513 INFO [main] org.apache.catalina.startup.VersionLoggerListener.log Command line argument: -Xmx256m`,
		},
	}
}

func (i *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:  inputName,
			Service: inputName,
			Pipeline: func() string {
				if i.Log != nil {
					return i.Log.Pipeline
				}
				return ""
			}(),
		},
	}
}

func (i *Input) RunPipeline() {
	l.Error("Collecting Tomcat in Jolokia way is deprecated. Exiting...")
}

func (i *Input) Run() {
	l.Error("Collecting Tomcat in Jolokia way is deprecated. Exiting...")
}

func (i *Input) Terminate() {
	if i.SemStop != nil { // nolint:typecheck
		i.SemStop.Close() // nolint:typecheck
	}
}

func defaultInput() *Input {
	return &Input{
		JolokiaAgent: inputs.JolokiaAgent{
			SemStop: cliutils.NewSem(),
			Tagger:  dkpt.DefaultGlobalTagger(),
		},
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
