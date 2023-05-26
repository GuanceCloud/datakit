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
	clipt "github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
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

const (
	inputName  = "mem"
	metricName = inputName
	sampleCfg  = `
[[inputs.mem]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'

[inputs.mem.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"`
)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	Interval     datakit.Duration
	Tags         map[string]string
	collectCache []inputs.Measurement

	vmStat   VMStat
	platform string

	semStop *cliutils.Sem // start stop signal
}

func (ipt *Input) Singleton() {
}

type memMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *memMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOpt())
}

// https://man7.org/linux/man-pages/man5/proc.5.html
// nolint:lll
func (m *memMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricName,
		Fields: map[string]interface{}{
			"total":             NewFieldInfoB("Total amount of memory."),
			"available":         NewFieldInfoB("Amount of available memory."),
			"available_percent": NewFieldInfoP("Available memory percent."),
			"used":              NewFieldInfoB("Amount of used memory."),
			"used_percent":      NewFieldInfoP("Used memory percent."),
			"active":            NewFieldInfoB("Memory that has been used more recently and usually not reclaimed unless absolutely necessary. (Darwin, Linux)"),
			"free":              NewFieldInfoB("Amount of free memory. (Darwin, Linux)"),
			"inactive":          NewFieldInfoB("Memory which has been less recently used.  It is more eligible to be reclaimed for other purposes. (Darwin, Linux)"),
			"wired":             NewFieldInfoB("Wired. (Darwin)"),
			"buffered":          NewFieldInfoB("Buffered. (Linux)"),
			"cached":            NewFieldInfoB("In-memory cache for files read from the disk. (Linux)"),
			"commit_limit":      NewFieldInfoB("This is the total amount of memory currently available to be allocated on the system. (Linux)"),
			"committed_as":      NewFieldInfoB("The amount of memory presently allocated on the system. (Linux)"),
			"dirty":             NewFieldInfoB("Memory which is waiting to get written back to the disk. (Linux)"),
			"high_free":         NewFieldInfoB("Amount of free high memory. (Linux)"),
			"high_total":        NewFieldInfoB("Total amount of high memory. (Linux)"),
			"huge_pages_free":   NewFieldInfoC("The number of huge pages in the pool that are not yet allocated. (Linux)"),
			"huge_pages_size":   NewFieldInfoB("The size of huge pages. (Linux)"),
			"huge_page_total":   NewFieldInfoC("The size of the pool of huge pages. (Linux)"),
			"low_free":          NewFieldInfoB("Amount of free low memory. (Linux)"),
			"low_total":         NewFieldInfoB("Total amount of low memory. (Linux)"),
			"mapped":            NewFieldInfoB("Files which have been mapped into memory, such as libraries. (Linux)"),
			"page_tables":       NewFieldInfoB("Amount of memory dedicated to the lowest level of page tables. (Linux)"),
			"shared":            NewFieldInfoB("Amount of shared memory. (Linux)"),
			"slab":              NewFieldInfoB("In-kernel data structures cache. (Linux)"),
			"sreclaimable":      NewFieldInfoB("Part of Slab, that might be reclaimed, such as caches. (Linux)"),
			"sunreclaim":        NewFieldInfoB("Part of Slab, that cannot be reclaimed on memory pressure. (Linux)"),
			"swap_cached":       NewFieldInfoB("Memory that once was swapped out, is swapped back in but still also is in the swap file. (Linux)"),
			"swap_free":         NewFieldInfoB("Amount of swap space that is currently unused. (Linux)"),
			"swap_total":        NewFieldInfoB("Total amount of swap space available. (Linux)"),
			"vmalloc_chunk":     NewFieldInfoB("Largest contiguous block of vmalloc area which is free. (Linux)"),
			"vmalloc_total":     NewFieldInfoB("Total size of vmalloc memory area. (Linux)"),
			"vmalloc_used":      NewFieldInfoB("Amount of vmalloc area which is used. (Linux)"),
			"write_back":        NewFieldInfoB("Memory which is actively being written back to the disk. (Linux)"),
			"write_back_tmp":    NewFieldInfoB("Memory used by FUSE for temporary write back buffers. (Linux)"),
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "System hostname."},
		},
	}
}

func (ipt *Input) Collect() error {
	ipt.collectCache = make([]inputs.Measurement, 0)
	vm, err := ipt.vmStat()
	if err != nil {
		return fmt.Errorf("error getting virtual memory info: %w", err)
	}

	fields := map[string]interface{}{
		"total":             vm.Total,
		"available":         vm.Available,
		"used":              vm.Used,
		"used_percent":      100 * float64(vm.Used) / float64(vm.Total),
		"available_percent": 100 * float64(vm.Available) / float64(vm.Total),
	}

	switch ipt.platform {
	case "darwin":
		fields["active"] = vm.Active
		fields["free"] = vm.Free
		fields["inactive"] = vm.Inactive
		fields["wired"] = vm.Wired
	case "linux":
		fields["active"] = vm.Active
		fields["buffered"] = vm.Buffers
		fields["cached"] = vm.Cached
		fields["commit_limit"] = vm.CommitLimit
		fields["committed_as"] = vm.CommittedAS
		fields["dirty"] = vm.Dirty
		fields["free"] = vm.Free
		fields["high_free"] = vm.HighFree
		fields["high_total"] = vm.HighTotal
		fields["huge_pages_free"] = vm.HugePagesFree
		fields["huge_pages_size"] = vm.HugePageSize
		fields["huge_pages_total"] = vm.HugePagesTotal
		fields["inactive"] = vm.Inactive
		fields["low_free"] = vm.LowFree
		fields["low_total"] = vm.LowTotal
		fields["mapped"] = vm.Mapped
		fields["page_tables"] = vm.PageTables
		fields["shared"] = vm.Shared
		fields["slab"] = vm.Slab
		fields["sreclaimable"] = vm.SReclaimable
		fields["sunreclaim"] = vm.SUnreclaim
		fields["swap_cached"] = vm.SwapCached
		fields["swap_free"] = vm.SwapFree
		fields["swap_total"] = vm.SwapTotal
		fields["vmalloc_chunk"] = vm.VMallocChunk
		fields["vmalloc_total"] = vm.VMallocTotal
		fields["vmalloc_used"] = vm.VMallocUsed
		fields["write_back_tmp"] = vm.WritebackTmp
		fields["write_back"] = vm.Writeback
	}
	tags := map[string]string{}
	for k, v := range ipt.Tags {
		tags[k] = v
	}
	ipt.collectCache = append(ipt.collectCache, &memMeasurement{
		name:   inputName,
		tags:   tags,
		fields: fields,
	})
	return err
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("memory input started")
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	for {
		start := time.Now()
		if err := ipt.Collect(); err != nil {
			l.Errorf("Collect: %s", err)
			io.FeedLastError(inputName, err.Error(), clipt.Metric)
		}

		if len(ipt.collectCache) > 0 {
			if err := inputs.FeedMeasurement(metricName, datakit.Metric, ipt.collectCache,
				&io.Option{CollectCost: time.Since(start)}); err != nil {
				l.Errorf("FeedMeasurement: %s", err)
			}
		}

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Infof("memory input exit")
			return

		case <-ipt.semStop.Wait():
			l.Infof("memory input return")
			return
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

func (*Input) SampleConfig() string {
	return sampleCfg
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&memMeasurement{},
	}
}

// ReadEnv support envs：
//   ENV_INPUT_MEM_TAGS : "a=b,c=d"
//   ENV_INPUT_MEM_INTERVAL : datakit.Duration
func (ipt *Input) ReadEnv(envs map[string]string) {
	if tagsStr, ok := envs["ENV_INPUT_MEM_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	//   ENV_INPUT_MEM_INTERVAL : datakit.Duration
	if str, ok := envs["ENV_INPUT_MEM_INTERVAL"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_MEM_INTERVAL to time.Duration: %s, ignore", err)
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
			platform: runtime.GOOS,
			vmStat:   VirtualMemoryStat,
			Interval: datakit.Duration{Duration: time.Second * 10},

			semStop: cliutils.NewSem(),
			Tags:    make(map[string]string),
		}
	})
}
