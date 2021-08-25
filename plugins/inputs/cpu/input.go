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
  ## Collect interval, default is 10 seconds. (optional)
  interval = '10s'
  ##
  ## Collect CPU usage per core, default is false. (optional)
  percpu = false
  ##
  ## Setting disable_temperature_collect to false will collect cpu temperature stats for linux.
  ##
  # disable_temperature_collect = false
  enable_temperature = true

  [inputs.cpu.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	PerCPU                    bool `toml:"percpu"`
	TotalCPU                  bool `toml:"totalcpu"`                    // deprecated
	CollectCPUTime            bool `toml:"collect_cpu_time"`            // deprecated
	ReportActive              bool `toml:"report_active"`               // deprecated
	DisableTemperatureCollect bool `toml:"disable_temperature_collect"` // deprecated

	EnableTemperature bool `toml:"enable_temperature"`

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
				Desc: "CPU core temperature. This is collected by default. Only collect the average temperature of all cores."},
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
	cpuTimes, err := i.ps.CPUTimes(i.PerCPU, true)
	if err != nil {
		return fmt.Errorf("error gather cpu stats: %s", err)
	}

	var coreTemp map[string]float64
	if !i.DisableTemperatureCollect || i.EnableTemperature {
		var errTemp error
		coreTemp, errTemp = CoreTemp()
		if errTemp != nil {
			l.Warn("failed to collect core temperature data: ", errTemp)
		}
	}

	time_now := time.Now()
	for _, cts := range cpuTimes {
		tags := map[string]string{
			"cpu": cts.CPU,
		}
		for k, v := range i.Tags {
			tags[k] = v
		}

		_, total := CpuActiveTotalTime(cts)

		lastCts, ok := i.lastStats[cts.CPU]
		if !ok {
			continue
		}
		_, lastTotal := CpuActiveTotalTime(lastCts)
		totalDelta := total - lastTotal
		if totalDelta < 0 {
			l.Error("error: current total cpu time less than previous total cpu time")
			break
		}
		if totalDelta == 0 {
			continue
		}
		cpuUsage, _ := CalculateUsage(cts, lastCts, totalDelta)

		fields := map[string]interface{}{
			"usage_user":       cpuUsage.User,
			"usage_system":     cpuUsage.System,
			"usage_idle":       cpuUsage.Idle,
			"usage_nice":       cpuUsage.Nice,
			"usage_iowait":     cpuUsage.Iowait,
			"usage_irq":        cpuUsage.Irq,
			"usage_softirq":    cpuUsage.Softirq,
			"usage_steal":      cpuUsage.Steal,
			"usage_guest":      cpuUsage.Guest,
			"usage_guest_nice": cpuUsage.GuestNice,
			"usage_total":      cpuUsage.Total,
		}

		if len(coreTemp) > 0 && cts.CPU == "cpu-total" {
			if v, ok := coreTemp[cts.CPU]; ok {
				fields["core_temperature"] = v
			}
		}
		i.appendMeasurement(inputName, tags, fields, time_now)

	}
	// last cputimes stats
	i.lastStats = make(map[string]cpu.TimesStat)
	for _, cts := range cpuTimes {
		i.lastStats[cts.CPU] = cts
	}
	return nil
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
			ps:                        &CPUInfo{},
			Interval:                  datakit.Duration{Duration: time.Second * 10},
			DisableTemperatureCollect: false,
			EnableTemperature:         false,
		}
	})
}
