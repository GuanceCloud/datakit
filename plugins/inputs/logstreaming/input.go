// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package logstreaming handle remote logging data.
package logstreaming

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	ihttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/workerpool"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	_ inputs.InputV2   = &Input{}
	_ inputs.HTTPInput = &Input{}
)

const (
	inputName        = "logstreaming"
	defaultPercision = "s"
	sampleCfg        = `
[inputs.logstreaming]
  ignore_url_tags = true

  ## Threads config controls how many goroutines an agent cloud start.
  ## buffer is the size of jobs' buffering of worker channel.
  ## threads is the total number fo goroutines at running time.
  # [inputs.logstreaming.threads]
    # buffer = 100
    # threads = 8
`
)

var (
	log   = logger.DefaultSLogger(inputName)
	wpool workerpool.WorkerPool
)

type Input struct {
	IgnoreURLTags bool                         `yaml:"ignore_url_tags"`
	WPConfig      *workerpool.WorkerPoolConfig `yaml:"threads"`
}

func (*Input) Catalog() string { return "log" }

func (*Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&logstreamingMeasurement{}}
}

func (ipt *Input) RegHTTPHandler() {
	log = logger.SLogger(inputName)

	if ipt.WPConfig != nil {
		wpool = workerpool.NewWorkerPool(ipt.WPConfig.Buffer)
		if err := wpool.Start(ipt.WPConfig.Threads); err != nil {
			log.Errorf("### start workerpool failed: %s", err.Error())
			wpool = nil
		}
	}

	dkhttp.RegHTTPHandler("POST", "/v1/write/logstreaming", ihttp.ProtectedHandlerFunc(ipt.handleLogstreaming, log))
}

func (*Input) Run() {
	log.Info("### register logstreaming router")
}

func (*Input) Terminate() {
	if wpool != nil {
		wpool.Shutdown()
		log.Debugf("### workerpool in %s is shudown", inputName)
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
