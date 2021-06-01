package cpu

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/cpu"
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
	inputName  = "cpu"
	metricName = inputName
	sampleCfg  = `
[[inputs.cpu]]
  # Collect interval, default is 10 seconds(optional)
  interval = '10s'

[inputs.cpu.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"`
)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	PerCPU         bool `toml:"percpu"`           // deprecated
	TotalCPU       bool `toml:"totalcpu"`         // deprecated
	CollectCPUTime bool `toml:"collect_cpu_time"` // deprecated
	ReportActive   bool `toml:"report_active"`    // deprecated

	Interval datakit.Duration
	Tags     map[string]string

	collectCache         []inputs.Measurement
	collectCacheLast1Ptr *cpuMeasurement

	lastStats map[string]cpu.TimesStat
	ps        CPUStatInfo
}

type cpuMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *cpuMeasurement) Info() *inputs.MeasurementInfo {
	// see https://man7.org/linux/man-pages/man5/proc.5.html
	return &inputs.MeasurementInfo{
		Name: metricName,
		Fields: map[string]interface{}{
			"usage_user": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "% CPU in user mode."},

			"usage_nice": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "% CPU in user mode with low priority (nice)."},

			"usage_system": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "% CPU in system mode."},

			"usage_idle": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "% CPU in the idle task."},

			"usage_iowait": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "% CPU waiting for I/O to complete."},

			"usage_irq": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "% CPU servicing hardware interrupts."},

			"usage_softirq": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "% CPU servicing soft interrupts."},

			"usage_steal": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "% CPU spent in other operating systems when running in a virtualized environment."},

			"usage_guest": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "% CPU spent running a virtual CPU for guest operating systems."},

			"usage_guest_nice": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "% CPU spent running a niced guest(virtual CPU for guest operating systems)."},

			"usage_total": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "% CPU in total active usage, as well as (100 - usage_idle)."},
			"core_temperature": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Celsius,
				Desc: "CPU core temperature"},
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "主机名"},
			"cpu":  &inputs.TagInfo{Desc: "CPU 核心"},
		},
	}
}

func (m *cpuMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (i *Input) appendMeasurement(name string, tags map[string]string, fields map[string]interface{}, ts time.Time) {
	tmp := &cpuMeasurement{name: name, tags: tags, fields: fields, ts: ts}
	i.collectCache = append(i.collectCache, tmp)
	i.collectCacheLast1Ptr = tmp
}

func (i *Input) Catalog() string {
	return "host"
}

func (i *Input) SampleConfig() string {
	return sampleCfg
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&cpuMeasurement{},
	}
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) Collect() error {
	// totalCPU only
	cpuTimes, err := i.ps.CPUTimes(false, true)
	if err != nil {
		return fmt.Errorf("error gather cpu stats: %s", err)
	}
	time_now := time.Now()
	for _, cts := range cpuTimes {
		tags := map[string]string{
			"cpu": cts.CPU,
		}
		for k, v := range i.Tags {
			tags[k] = v
		}
		fields := make(map[string]interface{})

		_, total := CpuActiveTotalTime(cts)

		lastCts, ok := i.lastStats[cts.CPU]
		if !ok {
			continue
		}
		_, lastTotal := CpuActiveTotalTime(lastCts)
		totalDelta := total - lastTotal
		if totalDelta < 0 {
			err = fmt.Errorf("error: current total cpu time less than previous total cpu time")
			break
		}
		if totalDelta == 0 {
			continue
		}
		cpuUsage, _ := CalculateUsage(cts, lastCts, totalDelta)

		if ok := CPUStatStructToMap(fields, cpuUsage, "usage_"); !ok {
			l.Error("error: collect cpu time, check cpu usage stat struct")
			break
		} else {
			if temp, err := CoreTempAvg(); err == nil {
				// 不增加新tag， 计算 core temp 的平均值
				fields["core_temperature"] = temp
			}
			i.appendMeasurement(inputName, tags, fields, time_now)
			// i.addField("active", 100 * (active-lastActive)/totalDelta)
		}
	}
	i.lastStats = make(map[string]cpu.TimesStat)
	for _, cts := range cpuTimes {
		i.lastStats[cts.CPU] = cts
	}
	return err
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("cpu input started")
	i.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, i.Interval.Duration)
	tick := time.NewTicker(i.Interval.Duration)
	isfirstRun := true
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			start := time.Now()
			if err := i.Collect(); err == nil {
				if errFeed := inputs.FeedMeasurement(metricName, datakit.Metric, i.collectCache,
					&io.Option{CollectCost: time.Since((start))}); errFeed != nil {
					if !isfirstRun {
						io.FeedLastError(inputName, errFeed.Error())
					} else {
						isfirstRun = false
					}
				}
			} else {
				io.FeedLastError(inputName, err.Error())
				l.Error(err)
			}
			i.collectCache = make([]inputs.Measurement, 0)
		case <-datakit.Exit.Wait():
			l.Infof("cpu input exit")
			return
		}
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			ps:       &CPUInfo{},
			Interval: datakit.Duration{Duration: time.Second * 10},
		}
	})
}
