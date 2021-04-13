package jvm

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = "jvm"
)

const (
	defaultInterval = "60s"
)

type Input struct {
	JolokiaAgent
}

type JvmMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (j *JvmMeasurement) LineProto() (*io.Point, error) {
	data, err := io.MakePoint(j.name, j.tags, j.fields, j.ts)
	fmt.Println(data.String())
	return data, err
}

func (j *JvmMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Fields: map[string]interface{}{
			"heap_memory_init":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The initial Java heap memory allocated."},
			"heap_memory_committed": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total Java heap memory committed to be used."},
			"heap_memory_max":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum Java heap memory available."},
			"heap_memory":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total Java heap memory used."},

			"non_heap_memory_init":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The initial Java non-heap memory allocated."},
			"non_heap_memory_committed": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total Java non-heap memory committed to be used."},
			"non_heap_memory_max":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum Java non-heap memory available."},
			"non_heap_memory":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total Java non-heap memory used."},

			"thread_count":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "The number of live threads."},
			"minor_collection_count": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "The number of minor garbage collections that have occurred."},
			"minor_collection_time":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The approximate minor garbage collection time elapsed."},
			"major_collection_count": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "The number of major garbage collections that have occurred."},
			"major_collection_time":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The approximate major garbage collection time elapsed."},
		},
	}
}

func (i *Input) Run() {
	if i.Interval == "" {
		i.Interval = defaultInterval
	}

	i.PluginName = inputName

	i.JolokiaAgent.Collect()
}

func (i *Input) Catalog() string      { return inputName }
func (i *Input) SampleConfig() string { return JvmConfigSample }
func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&JvmMeasurement{},
	}
}

func (i *Input) AvailableArchs() []string {
	return datakit.UnknownArch
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}

func (j *JolokiaAgent) Collect() {
	j.l = logger.DefaultSLogger(j.PluginName)
	j.l.Infof("%s input started...", j.PluginName)

	duration, err := time.ParseDuration(j.Interval)
	if err != nil {
		j.l.Error(err)
		return
	}

	tick := time.NewTicker(duration)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			start := time.Now()
			if err := j.Gather(); err != nil {
				j.l.Error(err)
			} else {
				inputs.FeedMeasurement(j.PluginName, io.Metric, j.collectCache,
					&io.Option{CollectCost: time.Since(start), HighFreq: false})

				j.collectCache = j.collectCache[:] // NOTE: do not forget to clean cache
			}

		case <-datakit.Exit.Wait():
			j.l.Infof("input %s exit", j.PluginName)
			return
		}
	}
}


