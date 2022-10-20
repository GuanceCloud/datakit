// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package rum real user monitoring
package rum

import (
	"net/http"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/workerpool"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	_ inputs.HTTPInput = &Input{}
	_ inputs.InputV2   = &Input{}
)

const (
	inputName    = "rum"
	sampleConfig = `
[[inputs.rum]]
  ## profile Agent endpoints register by version respectively.
  ## Endpoints can be skipped listen by remove them from the list.
  ## Default value set as below. DO NOT MODIFY THESE ENDPOINTS if not necessary.
    endpoints = ["/v1/write/rum"]

  ## Android command-line-tools HOME
    android_cmdline_home = "/usr/local/datakit/data/rum/tools/cmdline-tools"

  ## proguard HOME
    proguard_home = "/usr/local/datakit/data/rum/tools/proguard"

  ## android-ndk HOME
    ndk_home = "/usr/local/datakit/data/rum/tools/android-ndk"

  ## atos or atosl bin path
  ## for macOS datakit use the built-in tool atos default
  ## for Linux there are several tools that can be used to instead of macOS atos partially,
  ## such as https://github.com/everettjf/atosl-rs
    atos_bin_path = "/usr/local/datakit/data/rum/tools/atosl"

  ## Threads config controls how many goroutines an agent cloud start to handle HTTP request.
  ## buffer is the size of jobs' buffering of worker channel.
  ## threads is the total number fo goroutines at running time.
  # [inputs.rum.threads]
    # buffer = 100
    # threads = 8
`
)

var (
	log    = logger.DefaultSLogger(inputName)
	wkpool *workerpool.WorkerPool
)

type Input struct {
	Endpoints          []string                     `toml:"endpoints"`
	JavaHome           string                       `toml:"java_home"`
	AndroidCmdLineHome string                       `toml:"android_cmdline_home"`
	ProguardHome       string                       `toml:"proguard_home"`
	NDKHome            string                       `toml:"ndk_home"`
	AtosBinPath        string                       `toml:"atos_bin_path"`
	WPConfig           *workerpool.WorkerPoolConfig `toml:"threads"`
}

func (*Input) Catalog() string { return inputName }

func (*Input) AvailableArchs() []string { return datakit.AllOS }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&trace.TraceMeasurement{Name: inputName}}
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

	for _, endpoint := range ipt.Endpoints {
		dkhttp.RegHTTPHandler(http.MethodPost, endpoint, workerpool.HTTPWrapper(wkpool, ipt.handleRUM))
		log.Infof("### register RUM endpoint: %s", endpoint)
	}
}

func (ipt *Input) Run() {
	log.Infof("### RUM agent serving on: %+#v", ipt.Endpoints)

	loadSourcemapFile()
}

func (*Input) Terminate() {
	log.Info("### RUM exits")
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
