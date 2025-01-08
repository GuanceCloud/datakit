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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/ebpftrace/espan"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/ebpftrace/spans"
)

const (
	inputName   = "ebpftrace"
	catalogName = "ebpftrace"

	configSample = `
[[inputs.ebpftrace]]
  db_path = "%s"
  use_app_trace_id = true
  window = "20s"
  sampling_rate = 0.1
`
)

var (
	log           = logger.DefaultSLogger(inputName)
	defaultWindow = time.Second * 20
)

type Input struct {
	DBPath        string        `toml:"db_path"`
	SQLitePath    string        `toml:"sqlite_path"` // Deprecated
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

	httpapi.RemoveHTTPRoute("POST", "/v1/bpftracing")
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

	ulid, _ := espan.NewRandID()

	httpapi.RegHTTPRoute("POST", "/v1/bpftracing",
		apiBPFTracing(ulid, ipt.mrrunner))
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{}
}

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLabelLinux, datakit.LabelK8s}
}

var defaultExporter = func(pts []*point.Point) error {
	return dkio.DefaultFeeder().FeedV2(point.Tracing, pts,
		dkio.WithInputName("ebpf-tracing"))
}

func (ipt *Input) Run() {
	log = logger.SLogger(inputName)
	spans.Init()

	if ipt.mrrunner == nil {
		if !initMRRunner(ipt) {
			return
		}
	}
	ipt.mrrunner.Run(defaultExporter)
	select {
	case <-datakit.Exit.Wait():
		log.Info("ebpftrace input exit")
		return
	case <-ipt.semStop.Wait():
		log.Info("ebpftrace input exit")
		return
	}
}

func initMRRunner(ipt *Input) bool {
	if ipt.DBPath == "" {
		ipt.DBPath = filepath.Join(datakit.InstallDir, "ebpf_spandb/")
	}

	if err := os.MkdirAll(ipt.DBPath, os.ModePerm); err != nil {
		log.Error(err)
		return false
	}

	if err := NewMRRunner(ipt); err != nil {
		return false
	}

	return true
}

// DBPath    string        `toml:"db_path"`
// UseAppTraceID bool          `toml:"use_app_trace_id"`
// Window        time.Duration `toml:"window"`
// SamplingRate  float64       `toml:"sampling_rate"`

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{FieldName: "DBPath", ENVName: "DB_PATH", Type: doc.String, Example: "`/usr/local/datakit/ebpf_spandb/`", Desc: "SQLite database file storage path", DescZh: "数据库文件存放路径"},
		{FieldName: "UseAppTraceID", Type: doc.Boolean, Default: `false`, Desc: "Use application-side trace id instead of eBPF trace id", DescZh: "使用应用侧 trace id 替代 eBPF trace id"},
		{FieldName: "Window", Type: doc.TimeDuration, Default: `20s`, Desc: "Span's link time window", DescZh: "链路 span 的链接时间窗口"},
		{FieldName: "SamplingRate", Type: doc.Float, Example: `0.1`, Desc: "Link sampling rate", DescZh: "链路采样率"},
	}

	return doc.SetENVDoc("ENV_INPUT_EBPFTRACE_", infos)
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

	if v, ok := envs["ENV_INPUT_EBPFTRACE_DB_PATH"]; ok {
		ipt.DBPath = v
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
	Run(fn spans.Exporter)
	Stop()
}
