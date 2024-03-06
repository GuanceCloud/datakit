// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package mem collects host memory metrics.
package mem

import (
	"fmt"
	"runtime"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	minInterval = time.Second
	maxInterval = time.Minute
	inputName   = "mem"
	metricName  = inputName
)

var (
	_ inputs.ReadEnv   = (*Input)(nil)
	_ inputs.Singleton = (*Input)(nil)
	l                  = logger.DefaultSLogger(inputName)
)

type Input struct {
	Interval time.Duration
	Tags     map[string]string

	mergedTags map[string]string

	collectCache []*point.Point

	vmStat   VMStat
	platform string

	feeder dkio.Feeder
	tagger datakit.GlobalTagger

	semStop *cliutils.Sem
}

func (ipt *Input) Run() {
	ipt.setup()

	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	for {
		start := time.Now()
		if err := ipt.collect(); err != nil {
			l.Errorf("collect: %s", err)
			ipt.feeder.FeedLastError(err.Error(),
				dkio.WithLastErrorInput(inputName),
				dkio.WithLastErrorCategory(point.Metric),
			)
		}

		if len(ipt.collectCache) > 0 {
			if err := ipt.feeder.Feed(metricName, point.Metric, ipt.collectCache,
				&dkio.Option{CollectCost: time.Since(start)}); err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					dkio.WithLastErrorInput(inputName),
					dkio.WithLastErrorCategory(point.Metric),
				)
				l.Errorf("feed measurement: %s", err)
			}
		}

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Infof("%s input exit", inputName)
			return
		case <-ipt.semStop.Wait():
			l.Infof("%s input return", inputName)
			return
		}
	}
}

func (ipt *Input) setup() {
	l = logger.SLogger(inputName)
	l.Infof("memory input started")
	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)

	ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")
	l.Debugf("merged tags: %+#v", ipt.mergedTags)
}

func (ipt *Input) collect() error {
	ipt.collectCache = make([]*point.Point, 0)
	ts := time.Now()
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ts))

	vm, err := ipt.vmStat()
	if err != nil {
		return fmt.Errorf("error getting virtual memory info: %w", err)
	}

	var kvs point.KVs

	kvs = kvs.Add("total", vm.Total, false, false)
	kvs = kvs.Add("available", vm.Available, false, false)
	kvs = kvs.Add("used", vm.Used, false, false)
	kvs = kvs.Add("used_percent", 100*float64(vm.Used)/float64(vm.Total), false, false)
	kvs = kvs.Add("available_percent", 100*float64(vm.Available)/float64(vm.Total), false, false)

	switch ipt.platform {
	case "darwin":
		kvs = kvs.Add("active", vm.Active, false, false)
		kvs = kvs.Add("free", vm.Free, false, false)
		kvs = kvs.Add("inactive", vm.Inactive, false, false)
		kvs = kvs.Add("wired", vm.Wired, false, false)
	case "linux":
		kvs = kvs.Add("active", vm.Active, false, false)
		kvs = kvs.Add("buffered", vm.Buffers, false, false)
		kvs = kvs.Add("cached", vm.Cached, false, false)
		kvs = kvs.Add("commit_limit", vm.CommitLimit, false, false)
		kvs = kvs.Add("committed_as", vm.CommittedAS, false, false)
		kvs = kvs.Add("dirty", vm.Dirty, false, false)
		kvs = kvs.Add("free", vm.Free, false, false)
		kvs = kvs.Add("high_free", vm.HighFree, false, false)
		kvs = kvs.Add("high_total", vm.HighTotal, false, false)
		kvs = kvs.Add("huge_pages_free", vm.HugePagesFree, false, false)
		kvs = kvs.Add("huge_pages_size", vm.HugePageSize, false, false)
		kvs = kvs.Add("huge_pages_total", vm.HugePagesTotal, false, false)
		kvs = kvs.Add("inactive", vm.Inactive, false, false)
		kvs = kvs.Add("low_free", vm.LowFree, false, false)
		kvs = kvs.Add("low_total", vm.LowTotal, false, false)
		kvs = kvs.Add("mapped", vm.Mapped, false, false)
		kvs = kvs.Add("page_tables", vm.PageTables, false, false)
		kvs = kvs.Add("shared", vm.Shared, false, false)
		kvs = kvs.Add("slab", vm.Slab, false, false)
		kvs = kvs.Add("sreclaimable", vm.SReclaimable, false, false)
		kvs = kvs.Add("sunreclaim", vm.SUnreclaim, false, false)
		kvs = kvs.Add("swap_cached", vm.SwapCached, false, false)
		kvs = kvs.Add("swap_free", vm.SwapFree, false, false)
		kvs = kvs.Add("swap_total", vm.SwapTotal, false, false)
		kvs = kvs.Add("vmalloc_chunk", vm.VMallocChunk, false, false)
		kvs = kvs.Add("vmalloc_total", vm.VMallocTotal, false, false)
		kvs = kvs.Add("vmalloc_used", vm.VMallocUsed, false, false)
		kvs = kvs.Add("write_back_tmp", vm.WritebackTmp, false, false)
		kvs = kvs.Add("write_back", vm.Writeback, false, false)
	}

	for k, v := range ipt.mergedTags {
		kvs = kvs.AddTag(k, v)
	}

	ipt.collectCache = append(ipt.collectCache, point.NewPointV2(inputName, kvs, opts...))

	return nil
}

func (*Input) Singleton() {}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}
func (*Input) Catalog() string          { return "host" }
func (*Input) SampleConfig() string     { return sampleCfg }
func (*Input) AvailableArchs() []string { return datakit.AllOS }
func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&docMeasurement{},
	}
}

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{FieldName: "Interval"},
		{FieldName: "Tags"},
	}

	return doc.SetENVDoc("ENV_INPUT_MEM_", infos)
}

// ReadEnv support envs：
//
//	ENV_INPUT_MEM_TAGS : "a=b,c=d"
//	ENV_INPUT_MEM_INTERVAL : time.Duration
func (ipt *Input) ReadEnv(envs map[string]string) {
	if tagsStr, ok := envs["ENV_INPUT_MEM_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	//   ENV_INPUT_MEM_INTERVAL : time.Duration
	if str, ok := envs["ENV_INPUT_MEM_INTERVAL"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_MEM_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.Interval = config.ProtectedInterval(minInterval,
				maxInterval,
				da)
		}
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			feeder: dkio.DefaultFeeder(),
			tagger: datakit.DefaultGlobalTagger(),

			platform: runtime.GOOS,
			vmStat:   VirtualMemoryStat,
			Interval: time.Second * 10,

			semStop: cliutils.NewSem(),
			Tags:    make(map[string]string),
		}
	})
}
