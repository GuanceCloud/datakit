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

  ## Threads config controls how many goroutines an agent cloud start to handle HTTP request.
  ## buffer is the size of jobs' buffering of worker channel.
  ## threads is the total number fo goroutines at running time.
  # [inputs.logstreaming.threads]
    # buffer = 100
    # threads = 8
`
)

var (
	log    = logger.DefaultSLogger(inputName)
	wkpool *workerpool.WorkerPool
)

type Input struct {
	IgnoreURLTags bool                         `yaml:"ignore_url_tags"`
	WPConfig      *workerpool.WorkerPoolConfig `yaml:"threads"`
}

func (*Input) Catalog() string { return "log" }

func (*Input) AvailableArchs() []string { return datakit.AllOS }

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&logstreamingMeasurement{}}
}

func (ipt *Input) RegHTTPHandler() {
	log = logger.SLogger(inputName)

	if ipt.WPConfig != nil {
		var err error
		if wkpool, err = workerpool.NewWorkerPool(ipt.WPConfig, log); err != nil {
			log.Errorf("### new worker-pool failed: %s", err.Error())
		} else if err = wkpool.Start(); err != nil {
			log.Errorf("### start worker-pool failed: %s", err.Error())
			wkpool = nil
		}
	}

	dkhttp.RegHTTPHandler("POST", "/v1/write/logstreaming", ihttp.ProtectedHandlerFunc(ipt.handleLogstreaming, log))
}

func (*Input) Run() {
	log.Info("### register logstreaming router")
}

func (*Input) Terminate() {
	if wkpool != nil {
		wkpool.Shutdown()
		log.Debugf("### workerpool in %s is shudown", inputName)
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
