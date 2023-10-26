// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package ebpftrace connect span
package ebpftrace

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/ebpftrace/spans"
)

const (
	inputName   = "ebpftrace"
	catalogName = "ebpftrace"

	configSample = `
[[inputs.ebpftrace]]
  sqlite_path = "%s"
  use_app_trace_id = true
  window = "20s"
  sampling_rate = 0.1
`
)

var (
	l             = logger.DefaultSLogger(inputName)
	defaultWindow = time.Second * 20
)

type Input struct {
	SQLitePath    string        `toml:"sqlite_path"`
	UseAppTraceID bool          `toml:"use_app_trace_id"`
	Window        time.Duration `toml:"window"`
	SamplingRate  float64       `toml:"sampling_rate"`

	mrrunner MRRunnerInterface
	semStop  *cliutils.Sem // start stop signal
}

func (ipt *Input) Singleton() {}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) Catalog() string {
	return catalogName
}

func (ipt *Input) SampleConfig() string {
	return fmt.Sprintf(configSample,
		filepath.Join(datakit.InstallDir, "ebpf_spandb/"))
}

func (ipt *Input) RegHTTPHandler() {
	if ipt.mrrunner == nil {
		if !initMRRunner(ipt) {
			return
		}
	}

	ulid, _ := spans.NewULID()

	httpapi.RegHTTPRoute("POST", "/v1/bpftracing",
		apiBPFTracing(ulid, ipt.mrrunner))
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{}
}

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLabelLinux, datakit.LabelK8s}
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	spans.Init()

	if ipt.mrrunner == nil {
		if !initMRRunner(ipt) {
			return
		}
	}
	ipt.mrrunner.Run()
	select {
	case <-datakit.Exit.Wait():
		l.Info("ebpftrace input exit")
		return
	case <-ipt.semStop.Wait():
		l.Info("ebpftrace input exit")
		return
	}
}

func initMRRunner(ipt *Input) bool {
	if ipt.SQLitePath == "" {
		ipt.SQLitePath = filepath.Join(datakit.InstallDir, "ebpf_spandb/")
	}

	if err := os.MkdirAll(ipt.SQLitePath, os.ModePerm); err != nil {
		l.Error(err)
		return false
	}

	if err := NewMRRunner(ipt); err != nil {
		return false
	}

	return true
}

func (ipt *Input) ReadEnv(envs map[string]string) {
	if v, ok := envs["ENV_INPUT_EBPFTRACE_USE_APP_TRACE_ID"]; ok {
		switch v {
		case "", "f", "false", "FALSE", "False", "0":
			ipt.UseAppTraceID = false
		default:
			ipt.UseAppTraceID = true
		}
	} else {
		ipt.UseAppTraceID = true
	}

	if v, ok := envs["ENV_INPUT_EBPFTRACE_WINDOW"]; ok {
		if win, err := time.ParseDuration(v); err != nil {
			ipt.Window = defaultWindow
		} else {
			ipt.Window = win
		}
	} else {
		ipt.Window = defaultWindow
	}

	if v, ok := envs["ENV_INPUT_EBPFTRACE_SAMPLING_RATE"]; ok {
		ipt.SamplingRate, _ = strconv.ParseFloat(v, 64)
	}

	if v, ok := envs["ENV_INPUT_EBPFTRACE_SQLITE_PATH"]; ok {
		ipt.SQLitePath = v
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add("ebpftrace", func() inputs.Input {
		return &Input{
			semStop: cliutils.NewSem(),
		}
	})
}

type MRRunnerInterface interface {
	InsertSpans(pts []*point.Point)
	Run()
}
