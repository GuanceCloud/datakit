// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package swap collect host swap metrics.
package swap

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var (
	_ inputs.ReadEnv   = (*Input)(nil)
	_ inputs.Singleton = (*Input)(nil)
)

const (
	minInterval = time.Second
	maxInterval = time.Minute
)

var (
	inputName  = "swap"
	metricName = inputName
	l          = logger.DefaultSLogger(inputName)
	sampleCfg  = `
[[inputs.swap]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'
  ##

[inputs.swap.tags]
# some_tag = "some_value"
# more_tag = "some_other_value"

`
)

type Input struct {
	Interval             datakit.Duration
	Tags                 map[string]string
	collectCache         []*point.Point
	collectCacheLast1Ptr *point.Point
	swapStat             SwapStat

	semStop *cliutils.Sem // start stop signal
	feeder  dkio.Feeder
	Tagger  datakit.GlobalTagger
	ptsTime time.Time
}

func (ipt *Input) Singleton() {}

func (ipt *Input) appendMeasurement(name string, tags map[string]string, fields map[string]interface{}) {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime))

	pt := point.NewPoint(name,
		append(point.NewTags(tags), point.NewKVs(fields)...),
		opts...)

	ipt.collectCache = append(ipt.collectCache, pt)
	ipt.collectCacheLast1Ptr = pt
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (*Input) Catalog() string {
	return "host"
}

func (*Input) SampleConfig() string {
	return sampleCfg
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&swapMetric{},
	}
}

func (ipt *Input) Collect() error {
	if len(ipt.collectCache) > 0 {
		ipt.collectCache = ipt.collectCache[:0]
	}

	swap, err := ipt.swapStat()
	if err != nil {
		return fmt.Errorf("error getting swap memory info: %w", err)
	}

	fields := map[string]interface{}{
		"total":        swap.Total,
		"used":         swap.Used,
		"free":         swap.Free,
		"used_percent": swap.UsedPercent,
		"in":           swap.Sin,
		"out":          swap.Sout,
	}
	tags := map[string]string{}
	for k, v := range ipt.Tags {
		tags[k] = v
	}
	ipt.appendMeasurement(metricName, tags, fields)

	return nil
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("system input started")
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	ipt.ptsTime = ntp.Now()
	for {
		collectStart := time.Now()
		if err := ipt.Collect(); err == nil {
			if errFeed := ipt.feeder.Feed(
				point.Metric, ipt.collectCache,
				dkio.WithCollectCost(time.Since(collectStart)),
				dkio.WithSource(metricName),
			); errFeed != nil {
				ipt.feeder.FeedLastError(errFeed.Error(),
					metrics.WithLastErrorInput(inputName),
					metrics.WithLastErrorCategory(point.Metric),
				)
				l.Errorf("feed : %s", errFeed)
			}
		} else {
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
			)
			l.Error(err)
		}

		select {
		case tt := <-tick.C:
			ipt.ptsTime = inputs.AlignTime(tt, ipt.ptsTime, ipt.Interval.Duration)

		case <-datakit.Exit.Wait():
			l.Infof("system input exit")
			return

		case <-ipt.semStop.Wait():
			l.Infof("system input return")
			return
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{FieldName: "Interval"},
		{FieldName: "Tags"},
	}

	return doc.SetENVDoc("ENV_INPUT_SWAP_", infos)
}

// ReadEnv support envs：
//
//	ENV_INPUT_SWAP_TAGS : "a=b,c=d"
//	ENV_INPUT_SWAP_INTERVAL : datakit.Duration
func (ipt *Input) ReadEnv(envs map[string]string) {
	if tagsStr, ok := envs["ENV_INPUT_SWAP_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	//   ENV_INPUT_SWAP_INTERVAL : datakit.Duration
	if str, ok := envs["ENV_INPUT_SWAP_INTERVAL"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_SWAP_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.Interval.Duration = config.ProtectedInterval(minInterval,
				maxInterval,
				da)
		}
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			swapStat: PSSwapStat,
			Interval: datakit.Duration{Duration: time.Second * 10},

			semStop: cliutils.NewSem(),
			feeder:  dkio.DefaultFeeder(),
			Tagger:  datakit.DefaultGlobalTagger(),
			Tags:    make(map[string]string),
		}
	})
}
