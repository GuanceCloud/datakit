// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package dk collect Datakit metrics.
package dk

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
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

  # set false to disable Datakit self-metrics collection
  enabled = true

  # keep empty to collect all types(count/gauge/summary/...)
  metric_types = []

  # collect frequency
  interval = "30s"

  # Upload Datakit runtime profiles when resource thresholds are matched. Disabled by default.
  [inputs.dk.self_profiling]
    enabled  = false # enable threshold-triggered self profiling
    interval = "10s" # interval for checking process CPU and memory
    cooldown = "5m" # minimum interval between two profile collections

    # Profiles to collect on each trigger. CPU is sampled for duration/emergency_duration;
    # other profile types are collected as snapshots.
    enabled_types      = ["cpu", "heap", "goroutine"] # cpu, heap, goroutine
    duration           = "30s"                        # CPU sample duration for normal threshold triggers
    emergency_duration = "10s"                        # CPU sample duration for emergency threshold triggers

    # Resource bases for percent thresholds. Same unit style as [resource_limit].
    cpu_cores  = 2.0  # CPU cores used as the 100% base
    mem_max_mb = 4096 # memory MiB used as the 100% base

    # Normal thresholds use the average of the latest recent_points samples.
    recent_points     = 5     # number of samples for average thresholds
    cpu_usage_percent = 80    # average CPU percent threshold, based on cpu_cores; set 0 to disable this threshold
    mem_usage_percent = 80    # average memory percent threshold, based on mem_max_mb; set 0 to disable this threshold
    mem_usage_mb      = 3072  # average RSS memory threshold in MiB; set 0 to disable this threshold

    # Emergency thresholds use the current sample.
    mem_usage_percent_emergency = 95    # current memory percent threshold, based on mem_max_mb; set 0 to disable this threshold
    mem_usage_mb_emergency      = 0     # current RSS memory threshold in MiB; set 0 to disable this threshold

    # Local queue and upload settings. Profiles are queued locally before uploading to Dataway.
    cache_path        = "dk_self_profile" # disk queue path; relative path is under DataKit cache dir
    cache_capacity_mb = 1024              # disk queue capacity in MiB
    send_timeout      = "60s"             # timeout for each upload attempt
    send_retry_count  = 4                 # max upload attempts for each queued profile payload

[inputs.dk.tags]
   # tag1 = "val-1"
   # tag2 = "val-2"
`
	maxInterval = time.Minute
	minInterval = 5 * time.Second
)

type Input struct {
	Enabled bool `toml:"enabled"`

	MetricTypes []string          `toml:"metric_types"`
	Interval    time.Duration     `toml:"interval"`
	Tags        map[string]string `toml:"tags"`

	SelfProfiling *SelfProfilingConfig `toml:"self_profiling"`

	Tagger datakit.GlobalTagger `toml:"-"`
	feeder dkio.Feeder          `toml:"-"`

	url          string
	prom         *prom.Prom
	selfProfiler *selfProfiler
	selfProfileG *goroutine.Group
	semStop      *cliutils.Sem
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
			ENVName:   "INTERVAL",
			Type:      doc.TimeDuration,
			ConfField: "interval",
			Example:   "`10s`",
			Default:   "`30s`",
			Desc:      "Collect interval",
			DescZh:    "采集间隔",
		},
		{
			ENVName:   "ENABLE_SELF_PROFILING",
			Type:      doc.Boolean,
			ConfField: "self_profiling.enabled",
			Example:   "`true`",
			Default:   "`false`",
			Desc:      "Enable threshold-triggered DataKit self profiling",
			DescZh:    "开启 DataKit 自身阈值触发 Profile 采集",
		},
	}

	return doc.SetENVDoc("ENV_INPUT_DK_", infos)
}

// ReadEnv accept specific ENV settings to input.
//
//	ENV_INPUT_DK_INTERVAL(duration)
//	ENV_INPUT_DK_ENABLE_SELF_PROFILING(bool)
func (ipt *Input) ReadEnv(envs map[string]string) {
	if x := envs["ENV_INPUT_DK_INTERVAL"]; x != "" {
		if du, err := time.ParseDuration(x); err != nil {
			l.Warnf("parse ENV_INPUT_DK_INTERVAL %s failed: %s, ignored", x, err)
		} else {
			ipt.Interval = du
		}
	}

	if x, ok := envs["ENV_INPUT_DK_ENABLE_SELF_PROFILING"]; ok {
		ipt.setSelfProfilingEnabledFromEnv(x)
	}
}

func (ipt *Input) setSelfProfilingEnabledFromEnv(x string) {
	enabled, err := strconv.ParseBool(x)
	if err != nil {
		l.Warnf("parse ENV_INPUT_DK_ENABLE_SELF_PROFILING %s failed: %s, ignored", x, err)
		return
	}
	if ipt.SelfProfiling == nil {
		ipt.SelfProfiling = defaultSelfProfilingConfig()
	}
	ipt.SelfProfiling.Enabled = enabled
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

func (ipt *Input) startSelfProfiler() {
	if ipt.selfProfiler == nil {
		p, err := newSelfProfiler(ipt)
		if err != nil {
			l.Errorf("init datakit self profiling failed: %s", err)
			return
		}
		ipt.selfProfiler = p
	}

	if ipt.selfProfiler == nil {
		return
	}

	ipt.selfProfileG = goroutine.NewGroup(goroutine.Option{Name: "inputs_dk_self_profile"})
	ipt.selfProfileG.Go(func(ctx context.Context) error {
		ipt.selfProfiler.run(ctx)
		return nil
	})
}

func (ipt *Input) closeSelfProfiler() {
	if ipt.selfProfileG != nil {
		if err := ipt.selfProfileG.Wait(); err != nil {
			l.Warnf("wait datakit self profiling goroutine group failed: %s", err)
		}
		ipt.selfProfileG = nil
	}

	if ipt.selfProfiler != nil {
		ipt.selfProfiler.close()
		ipt.selfProfiler = nil
	}
}

func (ipt *Input) Run() {
	l = logger.SLogger(source)

	if !ipt.Enabled {
		l.Infof("%s input disabled", inputName)
		return
	}

	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)

	ipt.setup(config.Cfg.HTTPAPI.Listen)

	if !datakit.Docker {
		if x := datakit.GetEnv("ENV_INPUT_DK_ENABLE_SELF_PROFILING"); x != "" {
			ipt.setSelfProfilingEnabledFromEnv(x)
		}
	}

	ipt.startSelfProfiler()
	defer ipt.closeSelfProfiler()

	// init prom
	for {
		x, err := prom.NewProm(
			prom.WithLogger(l),
			prom.WithSource(source),
			prom.WithMetricTypes(ipt.MetricTypes),
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

	start := ntp.Now()
	for {
		collectStart := time.Now()
		pts, err := ipt.prom.CollectFromHTTPV2(ipt.url, prom.WithTimestamp(start.UnixNano()))
		if err != nil {
			l.Warnf("prom.CollectFromHTTPV2: %s, ignored", err.Error())
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
				metrics.WithLastErrorSource(source),
				metrics.WithLastErrorCategory(point.Metric),
			)
		} else if len(pts) > 0 {
			if err := ipt.feeder.Feed(point.Metric, pts,
				dkio.WithCollectCost(time.Since(collectStart)),
				dkio.WithElection(false),
				dkio.WithSource(source), dkio.WithInput(inputName)); err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					metrics.WithLastErrorInput(inputName),
					metrics.WithLastErrorCategory(point.Metric),
				)
				l.Errorf("feed measurement: %s", err)
			}
		}

		select {
		case tt := <-tick.C:
			start = inputs.AlignTime(tt, start, ipt.Interval)

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
		Enabled:  true,
		url:      fmt.Sprintf("http://%s/metrics", defaultHost),
		Interval: time.Second * 30,
		semStop:  cliutils.NewSem(),
		Tags:     map[string]string{},
		Tagger:   datakit.DefaultGlobalTagger(),

		SelfProfiling: defaultSelfProfilingConfig(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return def()
	})
}
