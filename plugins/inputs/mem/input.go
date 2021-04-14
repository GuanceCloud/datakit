package mem

import (
	"fmt"
	"runtime"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName    = "mem"
	metricName   = inputName
	collectCycle = time.Second * 10
	sampleCfg    = `
[[inputs.mem]]
# no sample need here, just open the input
`
)

type Input struct {
	collectCache []inputs.Measurement

	logger   *logger.Logger
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

// TODO
func (m *memMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricName,
		Fields: map[string]interface{}{
			"total":             NewFieldInfoB("total"),
			"available":         NewFieldInfoB("available"),
			"available_percent": NewFieldInfoP("available_percent"),
			"used":              NewFieldInfoB("used"),
			"used_percent":      NewFieldInfoP("used_percent"),

			"active":   NewFieldInfoB("active (Darwin, Linux)"),
			"free":     NewFieldInfoB("free (Darwin, Linux)"),
			"inactive": NewFieldInfoB("inactive (Darwin, Linux)"),
			"wired":    NewFieldInfoB("wired (Darwin)"),

			"buffered":        NewFieldInfoB("buffered (Linux)"),
			"cached":          NewFieldInfoB("cached (Linux)"),
			"commit_limit":    NewFieldInfoB("commit_limit (Linux)"),
			"committed_as":    NewFieldInfoB("committed_as (Linux)"),
			"dirty":           NewFieldInfoB("dirty (Linux)"),
			"high_free":       NewFieldInfoB("high_free (Linux)"),
			"high_total":      NewFieldInfoB("high_total (Linux)"),
			"huge_pages_free": NewFieldInfoB("huge_pages_free (Linux)"),
			"huge_page_size":  NewFieldInfoB("huge_page_size (Linux)"),
			"huge_page_total": NewFieldInfoB("huge_pages_total (Linux)"),
			"low_free":        NewFieldInfoB("low_free (Linux)"),
			"low_total":       NewFieldInfoB("low_total (Linux)"),
			"mapped":          NewFieldInfoB("mapped (Linux)"),
			"page_tables":     NewFieldInfoB("page_tables (Linux)"),
			"shared":          NewFieldInfoB("shared (Linux)"),
			"slab":            NewFieldInfoB("slab (Linux)"),
			"sreclaimable":    NewFieldInfoB("sreclaimable (Linux)"),
			"sunreclaim":      NewFieldInfoB("sunreclaim (Linux)"),
			"swap_cached":     NewFieldInfoB("swap_cached (Linux)"),
			"swap_free":       NewFieldInfoB("swap_free (Linux)"),
			"swap_total":      NewFieldInfoB("swap_total (Linux)"),
			"vmalloc_chunk":   NewFieldInfoB("vmalloc_chunk (Linux)"),
			"vmalloc_total":   NewFieldInfoB("vmalloc_total (Linux)"),
			"vmalloc_used":    NewFieldInfoB("vmalloc_used (Linux)"),
			"write_back":      NewFieldInfoB("write_back (Linux)"),
			"write_back_tmp":  NewFieldInfoB("write_back_tmp (Linux)"),
		},
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
		fields["huge_page_size"] = vm.HugePageSize
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
	i.collectCache = append(i.collectCache, &memMeasurement{
		name:   inputName,
		tags:   map[string]string{"host": datakit.Cfg.MainCfg.Hostname},
		fields: fields,
	})
	return err
}

func (i *Input) Run() {
	i.logger.Infof("memory input started")
	tick := time.NewTicker(collectCycle)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			start := time.Now()
			if err := i.Collect(); err == nil {
				inputs.FeedMeasurement(metricName, io.Metric, i.collectCache,
					&io.Option{CollectCost: time.Since(start)})
			} else {
				i.logger.Error(err)
			}
		case <-datakit.Exit.Wait():
			i.logger.Infof("memory input exit")
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
	return datakit.UnknownArch
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&memMeasurement{},
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			logger:   logger.SLogger(inputName),
			platform: runtime.GOOS,
			vmStat:   VirtualMemoryStat,
		}
	})
}
