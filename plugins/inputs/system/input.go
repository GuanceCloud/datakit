package system

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
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
	inputName  = "system"
	metricName = inputName
	sampleCfg  = `
[[inputs.system]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'

[inputs.system.tags]
# some_tag = "some_value"
# more_tag = "some_other_value"`
)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	Interval  datakit.Duration
	Fielddrop []string
	Tags      map[string]string

	collectCache         []inputs.Measurement
	collectCacheLast1Ptr *systemMeasurement
}

type systemMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (i *Input) appendMeasurement(name string, tags map[string]string, fields map[string]interface{}, ts time.Time) {
	tmp := &systemMeasurement{name: name, tags: tags, fields: fields, ts: ts}
	i.collectCache = append(i.collectCache, tmp)
	i.collectCacheLast1Ptr = tmp
}

func (i *Input) addField(field string, value interface{}) error {
	if i.collectCacheLast1Ptr == nil {
		return fmt.Errorf("error: append one before adding")
	}
	i.collectCacheLast1Ptr.fields[field] = value
	return nil
}

func (m *systemMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricName,
		Fields: map[string]interface{}{
			"load1":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "CPU load average over the last 1 minute"},
			"load5":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "CPU load average over the last 5 minutes"},
			"load15":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "CPU load average over the last 15 minutes"},
			"load1_per_core":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "CPU load average over the last 1 minute per core"},
			"load5_per_core":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "CPU load average over the last 5 minutes per core"},
			"load15_per_core": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "CPU load average over the last 15 minutes per core"},
			"n_cpus":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Count, Desc: "number of CPUs"},
			"n_users":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Count, Desc: "number of users"},
			"uptime":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "system uptime"},
			// "uptime_format": &inputs.FieldInfo{DataType: inputs.String, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "formatted system uptime"},
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "主机名"},
		},
	}
}

func (m *systemMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (i *Input) Catalog() string {
	return "host"
}

func (i *Input) SampleConfig() string {
	// 不记录 uptime_format
	// 配置文件中移除 `fielddrop = ["uptime_format"]`
	return sampleCfg
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&systemMeasurement{},
	}
}

func (i *Input) Collect() error {
	// clear collectCache
	i.collectCache = make([]inputs.Measurement, 0)
	loadAvg, err := load.Avg()
	if err != nil && !strings.Contains(err.Error(), "not implemented") {
		return err
	}
	numCPUs, err := cpu.Counts(true)
	if err != nil {
		return err
	}

	tags := map[string]string{}
	for k, v := range i.Tags {
		tags[k] = v
	}

	i.appendMeasurement(
		metricName,
		tags,
		map[string]interface{}{
			"load1":           loadAvg.Load1,
			"load5":           loadAvg.Load5,
			"load15":          loadAvg.Load15,
			"load1_per_core":  loadAvg.Load1 / float64(numCPUs),
			"load5_per_core":  loadAvg.Load5 / float64(numCPUs),
			"load15_per_core": loadAvg.Load15 / float64(numCPUs),
			"n_cpus":          numCPUs,
		},
		time.Now(),
	)

	users, err := host.Users()
	if err == nil {
		i.addField("n_users", len(users))
	} else if os.IsNotExist(err) {
		l.Debugf("Reading users: %s", err.Error())
	} else if os.IsPermission(err) {
		l.Debug(err.Error())
	}
	uptime, err := host.Uptime()
	if err == nil {
		i.addField("uptime", uptime)
	}

	return err
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("system input started")
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
				// i.collectCache = make([]inputs.Measurement, 0)
			} else {
				io.FeedLastError(inputName, err.Error())
				l.Error(err)
			}
		case <-datakit.Exit.Wait():
			l.Infof("system input exit")
			return
		}
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Interval: datakit.Duration{Duration: time.Second * 10},
		}
	})
}
