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
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
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

type swapMeasurement struct{}

func (m *swapMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricName,
		Type: "metric",
		Fields: map[string]interface{}{
			"total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Host swap memory free.",
			},
			"used": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Host swap memory used.",
			},
			"free": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Host swap memory total.",
			},
			"used_percent": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "Host swap memory percentage used.",
			},
			"in": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Moving data from swap space to main memory of the machine.",
			},
			"out": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Moving main memory contents to swap disk when main memory space fills up.",
			},
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "hostname"},
		},
	}
}

type Input struct {
	Interval             datakit.Duration
	Tags                 map[string]string
	collectCache         []*point.Point
	collectCacheLast1Ptr *point.Point
	swapStat             SwapStat

	semStop *cliutils.Sem // start stop signal
	feeder  dkio.Feeder
	Tagger  datakit.GlobalTagger
}

func (ipt *Input) Singleton() {
}

func (ipt *Input) appendMeasurement(name string, tags map[string]string, fields map[string]interface{}, ts time.Time) {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ts))

	pt := point.NewPointV2(name,
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
		&swapMeasurement{},
	}
}

func (ipt *Input) Collect() error {
	ipt.collectCache = make([]*point.Point, 0)
	swap, err := ipt.swapStat()
	ts := time.Now()
	if err != nil {
		return fmt.Errorf("error getting swap memory info: %w", err)
	}

	fields := map[string]interface{}{
		"total":        swap.Total,
		"used":         swap.Used,
		"free":         swap.Free,
		"used_percent": swap.UsedPercent,

		"in":  swap.Sin,
		"out": swap.Sout,
	}
	tags := map[string]string{}
	for k, v := range ipt.Tags {
		tags[k] = v
	}
	ipt.appendMeasurement(metricName, tags, fields, ts)

	return nil
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("system input started")
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			start := time.Now()
			if err := ipt.Collect(); err == nil {
				if errFeed := ipt.feeder.Feed(metricName, point.Metric, ipt.collectCache,
					&dkio.Option{CollectCost: time.Since(start)}); errFeed != nil {
					dkio.FeedLastError(inputName, errFeed.Error())
					l.Error(errFeed)
				}
			} else {
				dkio.FeedLastError(inputName, err.Error())
				l.Error(err)
			}
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
