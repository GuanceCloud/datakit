// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package dk collect Datakit metrics.
package dk

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
)

const (
	inputName   = "dk"
	source      = "dk-metrics"
	measurement = "dk"
)

var (
	l            = logger.DefaultSLogger(source)
	defaultHost  = "localhost:9529"
	configSample = `
[[inputs.dk]]

  # See https://docs.<<<custom_key.brand_main_domain>>>/datakit/datakit-metrics/#metrics for all metrics exported by Datakit.
  metric_name_filter = [
    ### Collect all metrics(these may collect 300+ metrics of Datakit)
    ### if you want to collect all, make this rule the first in the list.
    # ".*",

    "datakit_http.*",       # HTTP API
    "datakit_goroutine.*",  # Goroutine

    ### runtime related
    "datakit_cpu_.*",
    "datakit_.*_alloc_bytes", # Memory
    "datakit_open_files",
    "datakit_uptime_seconds",
    "datakit_data_overuse",
    "datakit_process_.*",

    ### election
    "datakit_election_status",

    ### Dataway related
    #"datakit_io_dataway_.*",
    #"datakit_io_http_retry_total",

    ### Filter
    #"datakit_filter_.*",

    ### dialtesting
    #"datakit_dialtesting_.*",

    ### Input feed
    #".*_feed_.*",
  ]

  # keep empty to collect all types(count/gauge/summary/...)
  metric_types = []

  # collect frequency
  interval = "30s"

[inputs.dk.tags]
   # tag1 = "val-1"
   # tag2 = "val-2"
`
	maxInterval = time.Minute
	minInterval = 5 * time.Second
)

type Input struct {
	MetricFilter []string          `toml:"metric_name_filter"`
	MetricTypes  []string          `toml:"metric_types"`
	Interval     time.Duration     `toml:"interval"`
	Tags         map[string]string `toml:"tags"`

	Tagger datakit.GlobalTagger `toml:"-"`
	feeder dkio.Feeder          `toml:"-"`

	url     string
	prom    *prom.Prom
	semStop *cliutils.Sem
}

// Singleton make the input only 1 instance when multiple instance configured.
func (*Input) Singleton() {}

// We should block these metrics to upload to workerspace, this may eat
// too many time series.
var alwaysBlockedMetrics = []string{
	metrics.DatakitLastError,
}

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	infos := []*inputs.ENVInfo{
		{
			ENVName:   "ENABLE_ALL_METRICS",
			Type:      doc.Boolean,
			Example:   `true`,
			Default:   doc.NoDefaultSet,
			ConfField: doc.NoField,
			Desc:      "Collect all metrics, any string",
			DescZh:    "采集所有指标，任意非空字符串",
		},
		{
			ENVName:   "ADD_METRICS",
			Type:      doc.List,
			Default:   doc.NoDefaultSet,
			ConfField: doc.NoField,
			Example:   "`[\"datakit_io_.*\", \"datakit_pipeline_.*\"]`",
			Desc:      "Additional metrics, Available metrics list [here](../datakit/datakit-metrics.md)",
			DescZh:    "追加指标列表，可用的指标名参见[这里](../datakit/datakit-metrics.md)",
		},

		{
			ENVName:   "ONLY_METRICS",
			Type:      doc.List,
			ConfField: doc.NoField,
			Example:   "`[\"datakit_io_.*\", \"datakit_pipeline_.*\"]`",
			Default:   doc.NoDefaultSet,
			Desc:      "Only enable metrics",
			DescZh:    "只开启指定指标",
		},
	}

	return doc.SetENVDoc("ENV_INPUT_DK_", infos)
}

// ReadEnv accept specific ENV settings to input.
//
//	ENV_INPUT_DK_ENABLE_ALL_METRICS(bool)
//	ENV_INPUT_DK_ADD_METRICS(json-string-list)
//	ENV_INPUT_DK_ONLY_METRICS(json-string-list)
func (ipt *Input) ReadEnv(envs map[string]string) {
	if _, ok := envs["ENV_INPUT_DK_ENABLE_ALL_METRICS"]; ok {
		ipt.MetricFilter = nil
	}

	if x := envs["ENV_INPUT_DK_ADD_METRICS"]; x != "" {
		arr := []string{}
		if err := json.Unmarshal([]byte(x), &arr); err != nil {
			l.Warnf("json.Unmarshal: %s, ignored", err)
		} else {
			ipt.MetricFilter = append(ipt.MetricFilter, arr...)
		}
	}

	if x := envs["ENV_INPUT_DK_ONLY_METRICS"]; x != "" {
		arr := []string{}
		if err := json.Unmarshal([]byte(x), &arr); err != nil {
			l.Warnf("json.Unmarshal: %s, ignored", err)
		} else {
			ipt.MetricFilter = arr
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (*Input) Catalog() string {
	return "host"
}

func (ipt *Input) SampleConfig() string {
	return configSample
}

func (ipt *Input) setup(listen string) {
	// setup tags
	for k, v := range ipt.Tagger.HostTags() {
		if _, ok := ipt.Tags[k]; !ok { // add global tags if not exist.
			l.Infof("add global tag %q:%q", k, v)
			ipt.Tags[k] = v
		}
	}

	// 'url' not config.Cfg.HTTPAPI.Listen, we force redirect to current listen address
	if u, err := url.Parse(ipt.url); err == nil {
		if u.Host != listen {
			l.Infof("force redirect URL from %q to %q", u.Host, listen)
			u.Host = listen
			ipt.url = u.String()
		}
	}
}

func (ipt *Input) Run() {
	l = logger.SLogger(source)

	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)

	ipt.setup(config.Cfg.HTTPAPI.Listen)

	// init prom
	for {
		x, err := prom.NewProm(
			prom.WithLogger(l),
			prom.WithSource(source),
			prom.WithMetricTypes(ipt.MetricTypes),
			prom.WithMetricNameFilter(ipt.MetricFilter),
			prom.WithMetricNameFilterIgnore(alwaysBlockedMetrics),
			prom.WithMeasurementName(measurement),
			prom.WithTags(ipt.Tags),
		)

		if err != nil {
			l.Errorf("prom.NewProm: %s", err)
			select {
			case <-datakit.Exit.Wait():
				return
			case <-ipt.semStop.Wait():
				l.Infof("%s input return", inputName)
				return
			default:
				time.Sleep(time.Second)
			}
		} else {
			ipt.prom = x
			break
		}
	}

	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	start := time.Now()
	for {
		pts, err := ipt.prom.CollectFromHTTPV2(ipt.url, prom.WithTimestamp(start.UnixNano()))
		if err != nil {
			l.Warnf("prom.CollectFromHTTPV2: %s, ignored", err.Error())
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
				metrics.WithLastErrorSource(source),
				metrics.WithLastErrorCategory(point.Metric),
			)
		} else if len(pts) > 0 {
			if err := ipt.feeder.FeedV2(point.Metric, pts,
				dkio.WithCollectCost(time.Since(start)),
				dkio.WithElection(false),
				dkio.WithInputName(source)); err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					metrics.WithLastErrorInput(inputName),
					metrics.WithLastErrorCategory(point.Metric),
				)
				l.Errorf("feed measurement: %s", err)
			}
		}

		select {
		case tt := <-tick.C:
			start = time.UnixMilli(inputs.AlignTimeMillSec(tt, start.UnixMilli(), ipt.Interval.Milliseconds()))

		case <-datakit.Exit.Wait():
			return

		case <-ipt.semStop.Wait():
			l.Infof("%s input return", inputName)
			return
		}
	}
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOS
}

func def() *Input {
	return &Input{
		feeder:   dkio.DefaultFeeder(),
		url:      fmt.Sprintf("http://%s/metrics", defaultHost),
		Interval: time.Second * 30,
		semStop:  cliutils.NewSem(),
		Tags:     map[string]string{},
		Tagger:   datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return def()
	})
}
