// Package output handle multiple output data.
package output

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "output"

	sampleCfg = `
[inputs.output]
  ignore_url_tags = true
`
)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	IgnoreURLTags bool `yaml:"ignore_url_tags"`
}

func (*Input) Catalog() string { return "log" }

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) AvailableArchs() []string { return datakit.AllArch }

// func (*Input) SampleMeasurement() []inputs.Measurement {
// 	return []inputs.Measurement{&logstreamingMeasurement{}}
// }

func (*Input) Run() {
	l.Info("register logstreaming router")
}

func (ipt *Input) RegHTTPHandler() {
	// l = logger.SLogger(inputName)
	// dhttp.RegHTTPHandler("POST", "/v1/write/logstreaming", ihttp.ProtectedHandlerFunc(i.handleLogstreaming, l))
}

func (ipt *Input) Terminate() {
	//
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
