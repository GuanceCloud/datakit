// Package cpu collect CPU metrics.
package cpu

import (
	"fmt"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var _ inputs.ReadEnv = (*Input)(nil)

const (
	minInterval = time.Second
	maxInterval = time.Minute
)

const (
	inputName  = "cpu"
	metricName = inputName
)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	TotalCPU                  bool `toml:"totalcpu"`                              // deprecated
	CollectCPUTime            bool `toml:"collect_cpu_time"`                      // deprecated
	ReportActive              bool `toml:"report_active"`                         // deprecated
	DisableTemperatureCollect bool `toml:"disable_temperature_collect,omitempty"` // deprecated

	PerCPU            bool `toml:"percpu"`
	EnableTemperature bool `toml:"enable_temperature"`

	Interval datakit.Duration
	Tags     map[string]string

	collectCache         []inputs.Measurement
	collectCacheLast1Ptr *cpuMeasurement

	lastStats map[string]cpu.TimesStat
	ps        CPUStatInfo
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
		return fmt.Errorf("error gather cpu stats: %w", err)
	}

	var coreTemp map[string]float64
	if i.EnableTemperature {
		var errTemp error
		coreTemp, errTemp = CoreTemp()
		if errTemp != nil {
			l.Warn("failed to collect core temperature data: ", errTemp)
			i.EnableTemperature = false
			l.Warn("skip core temperature data collection")
		}
	}

	now := time.Now()
	for _, cts := range cpuTimes {
		tags := map[string]string{
			"cpu": cts.CPU,
		}
		for k, v := range i.Tags {
			tags[k] = v
		}

		_, total := CPUActiveTotalTime(cts)

		lastCts, ok := i.lastStats[cts.CPU]
		if !ok {
			continue
		}
		_, lastTotal := CPUActiveTotalTime(lastCts)
		totalDelta := total - lastTotal
		if totalDelta < 0 {
			l.Warn("error: current total cpu time less than previous total cpu time")
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
		i.appendMeasurement(inputName, tags, fields, now)
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

	if err := i.Collect(); err != nil { // gather lastSats
		l.Errorf("Collect: %s", err.Error())
	}

	time.Sleep(i.Interval.Duration)

	tick := time.NewTicker(i.Interval.Duration)
	defer tick.Stop()

	for {
		start := time.Now()
		if err := i.Collect(); err != nil {
			l.Errorf("Collect: %s", err)
			io.FeedLastError(inputName, err.Error())
		}

		if len(i.collectCache) > 0 {
			if err := inputs.FeedMeasurement(metricName, datakit.Metric, i.collectCache,
				&io.Option{CollectCost: time.Since((start))}); err != nil {
				l.Errorf("FeedMeasurement: %s", err)
			}
		}
		i.collectCache = make([]inputs.Measurement, 0)

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Infof("cpu input exit")
			return
		}
	}
}

// ReadEnv support envs：
//   ENV_INPUT_CPU_PERCPU : booler
//   ENV_INPUT_CPU_ENABLE_TEMPERATURE : booler
//   ENV_INPUT_CPU_TAGS : "a=b,c=d"
func (i *Input) ReadEnv(envs map[string]string) {
	if percpu, ok := envs["ENV_INPUT_CPU_PERCPU"]; ok {
		b, err := strconv.ParseBool(percpu)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CPU_PERCPU to bool: %s, ignore", err)
		} else {
			i.PerCPU = b
		}
	}

	if enableTemperature, ok := envs["ENV_INPUT_CPU_ENABLE_TEMPERATURE"]; ok {
		b, err := strconv.ParseBool(enableTemperature)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CPU_ENABLE_TEMPERATURE to bool: %s, ignore", err)
		} else {
			i.EnableTemperature = b
		}
	}

	if tagsStr, ok := envs["ENV_INPUT_CPU_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			i.Tags[k] = v
		}
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			ps:                &CPUInfo{},
			Interval:          datakit.Duration{Duration: time.Second * 10},
			EnableTemperature: true,
			Tags:              make(map[string]string),
		}
	})
}
