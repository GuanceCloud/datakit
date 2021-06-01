package mem

import (
	"fmt"
	"runtime"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
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
}

type memMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *memMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

// https://man7.org/linux/man-pages/man5/proc.5.html

func (m *memMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricName,
		Fields: map[string]interface{}{
			"total":             NewFieldInfoB("Total amount of memory"),
			"available":         NewFieldInfoB("Amount of available memory"),
			"available_percent": NewFieldInfoP("Available memory percent"),
			"used":              NewFieldInfoB("Amount of used memory"),
			"used_percent":      NewFieldInfoP("Used memory percent"),
			"active": NewFieldInfoB("Memory that has been used more recently and usually " +
				"not reclaimed unless absolutely necessary. (Darwin, Linux)"),
			"free": NewFieldInfoB("Amount of free memory(Darwin, Linux)"),
			"inactive": NewFieldInfoB("Memory which has been less recently used.  It is " +
				"more eligible to be reclaimed for other purposes. (Darwin, Linux)"),
			"wired":    NewFieldInfoB("wired (Darwin)"),
			"buffered": NewFieldInfoB("buffered (Linux)"),
			"cached":   NewFieldInfoB("In-memory cache for files read from the disk. (Linux)"),
			"commit_limit": NewFieldInfoB("This is the total amount of memory currently " +
				"available to be allocated on the system. (Linux)"),
			"committed_as":    NewFieldInfoB("The amount of memory presently allocated on the system. (Linux)"),
			"dirty":           NewFieldInfoB("Memory which is waiting to get written back to the disk. (Linux)"),
			"high_free":       NewFieldInfoB("Amount of free highmem. (Linux)"),
			"high_total":      NewFieldInfoB("Total amount of highmem. (Linux)"),
			"huge_pages_free": NewFieldInfoC("The number of huge pages in the pool that are not yet allocated. (Linux)"),
			"huge_pages_size": NewFieldInfoB("The size of huge pages. (Linux)"),
			"huge_page_total": NewFieldInfoC("The size of the pool of huge pages. (Linux)"),
			"low_free":        NewFieldInfoB("Amount of free lowmem. (Linux)"),
			"low_total":       NewFieldInfoB("Total amount of lowmem. (Linux)"),
			"mapped":          NewFieldInfoB("Files which have been mapped into memory, such as libraries. (Linux)"),
			"page_tables":     NewFieldInfoB("Amount of memory dedicated to the lowest level of page tables. (Linux)"),
			"shared":          NewFieldInfoB("Amount of shared memory (Linux)"),
			"slab":            NewFieldInfoB("In-kernel data structures cache. (Linux)"),
			"sreclaimable":    NewFieldInfoB("Part of Slab, that might be reclaimed, such as caches. (Linux)"),
			"sunreclaim":      NewFieldInfoB("Part of Slab, that cannot be reclaimed on memory pressure. (Linux)"),
			"swap_cached":     NewFieldInfoB("Memory that once was swapped out, is swapped back in but still also is in the swap file. (Linux)"),
			"swap_free":       NewFieldInfoB("Amount of swap space that is currently unused. (Linux)"),
			"swap_total":      NewFieldInfoB("Total amount of swap space available. (Linux)"),
			"vmalloc_chunk":   NewFieldInfoB("Largest contiguous block of vmalloc area which is free. (Linux)"),
			"vmalloc_total":   NewFieldInfoB("Total size of vmalloc memory area. (Linux)"),
			"vmalloc_used":    NewFieldInfoB("Amount of vmalloc area which is used. (Linux)"),
			"write_back":      NewFieldInfoB("Memory which is actively being written back to the disk. (Linux)"),
			"write_back_tmp":  NewFieldInfoB("Memory used by FUSE for temporary writeback buffers. (Linux)"),
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "主机名"}},
	}
}

func (i *Input) Collect() error {
	i.collectCache = make([]inputs.Measurement, 0)
	vm, err := i.vmStat()
	if err != nil {
		return fmt.Errorf("error getting virtual memory info: %s", err)
	}

	fields := map[string]interface{}{
		"total":             vm.Total,
		"available":         vm.Available,
		"used":              vm.Used,
		"used_percent":      100 * float64(vm.Used) / float64(vm.Total),
		"available_percent": 100 * float64(vm.Available) / float64(vm.Total),
	}

	switch i.platform {
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
	for k, v := range i.Tags {
		tags[k] = v
	}
	i.collectCache = append(i.collectCache, &memMeasurement{
		name:   inputName,
		tags:   tags,
		fields: fields,
	})
	return err
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("memory input started")
	i.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, i.Interval.Duration)
	tick := time.NewTicker(i.Interval.Duration)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			start := time.Now()
			if err := i.Collect(); err == nil {
				if errFeed := inputs.FeedMeasurement(metricName, datakit.Metric, i.collectCache,
					&io.Option{CollectCost: time.Since(start)}); errFeed != nil {
					io.FeedLastError(inputName, errFeed.Error())
					l.Error(errFeed)
				}
			} else {
				io.FeedLastError(inputName, err.Error())
				l.Error(err)
			}
		case <-datakit.Exit.Wait():
			l.Infof("memory input exit")
			return
		}
	}
}

func (i *Input) Catalog() string {
	return "host"
}

func (i *Input) SampleConfig() string {
	return sampleCfg
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&memMeasurement{},
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			platform: runtime.GOOS,
			vmStat:   VirtualMemoryStat,
			Interval: datakit.Duration{Duration: time.Second * 10},
		}
	})
}
