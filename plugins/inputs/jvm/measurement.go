//nolint:lll
package jvm

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type JvmMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

type JavaRuntimeMemt struct {
	JvmMeasurement
}

type JavaMemoryMemt struct {
	JvmMeasurement
}

type JavaGcMemt struct {
	JvmMeasurement
}

type JavaLastGcMemt struct {
	JvmMeasurement
}

type JavaThreadMemt struct {
	JvmMeasurement
}

type JavaClassLoadMemt struct {
	JvmMeasurement
}

type JavaMemoryPoolMemt struct {
	JvmMeasurement
}

func (j *JvmMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(j.name, j.tags, j.fields, j.ts)
}

func (j *JvmMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{}
}

func (j *JavaRuntimeMemt) LineProto() (*io.Point, error) {
	return io.MakePoint(j.name, j.tags, j.fields, j.ts)
}

func (j *JavaRuntimeMemt) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "java_runtime",
		Fields: map[string]interface{}{
			"Uptime": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "The total runtime."},
		},
		Tags: map[string]interface{}{
			"jolokia_agent_url": inputs.TagInfo{Desc: "jolokia agent url path"},
		},
	}
}

func (j *JavaMemoryMemt) LineProto() (*io.Point, error) {
	return io.MakePoint(j.name, j.tags, j.fields, j.ts)
}

func (j *JavaMemoryMemt) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "java_memory",
		Fields: map[string]interface{}{
			"HeapMemoryUsageinit":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The initial Java heap memory allocated."},
			"HeapMemoryUsageused":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total Java heap memory used."},
			"HeapMemoryUsagemax":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum Java heap memory available."},
			"HeapMemoryUsagecommitted": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total Java heap memory committed to be used."},

			"NonHeapMemoryUsageinit":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The initial Java non-heap memory allocated."},
			"NonHeapMemoryUsageused":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total Java non-heap memory used."},
			"NonHeapMemoryUsagemax":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum Java non-heap memory available."},
			"NonHeapMemoryUsagecommitted": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total Java non-heap memory committed to be used."},

			"ObjectPendingFinalizationCount": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The count of object pending finalization."},
		},

		Tags: map[string]interface{}{
			"jolokia_agent_url": inputs.NewTagInfo("jolokia agent url path"),
		},
	}
}

func (j *JavaGcMemt) LineProto() (*io.Point, error) {
	return io.MakePoint(j.name, j.tags, j.fields, j.ts)
}

func (j *JavaGcMemt) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "java_garbage_collector",
		Fields: map[string]interface{}{
			"CollectionTime":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The approximate GC collection time elapsed."},
			"CollectionCount": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of GC that have occurred."},
		},

		Tags: map[string]interface{}{
			"jolokia_agent_url": inputs.NewTagInfo("jolokia agent url path"),
			"name":              inputs.NewTagInfo("the name of GC generation"),
		},
	}
}

func (j *JavaLastGcMemt) LineProto() (*io.Point, error) {
	return io.MakePoint(j.name, j.tags, j.fields, j.ts)
}

func (j *JavaLastGcMemt) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{}
}

func (j *JavaThreadMemt) LineProto() (*io.Point, error) {
	return io.MakePoint(j.name, j.tags, j.fields, j.ts)
}

func (j *JavaThreadMemt) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "java_threading",
		Fields: map[string]interface{}{
			"DaemonThreadCount":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The count of daemon thread."},
			"PeakThreadCount":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The peak count of thread."},
			"ThreadCount":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The count of thread."},
			"TotalStartedThreadCount": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total count of started thread."},
		},

		Tags: map[string]interface{}{
			"jolokia_agent_url": inputs.NewTagInfo("jolokia agent url path"),
		},
	}
}

func (j *JavaClassLoadMemt) LineProto() (*io.Point, error) {
	return io.MakePoint(j.name, j.tags, j.fields, j.ts)
}

func (j *JavaClassLoadMemt) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "java_class_loading",
		Fields: map[string]interface{}{
			"LoadedClassCount":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The count of loaded class."},
			"TotalLoadedClassCount": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total count of loaded class."},
			"UnloadedClassCount":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The count of unloaded class."},
		},

		Tags: map[string]interface{}{
			"jolokia_agent_url": inputs.NewTagInfo("jolokia agent url path"),
		},
	}
}

func (j *JavaMemoryPoolMemt) LineProto() (*io.Point, error) {
	return io.MakePoint(j.name, j.tags, j.fields, j.ts)
}

func (j *JavaMemoryPoolMemt) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "java_memory_pool",
		Fields: map[string]interface{}{
			"Usageinit":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The initial Java memory pool allocated"},
			"Usagemax":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum Java  memory pool available."},
			"Usagecommitted": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total Java memory pool committed to be used"},
			"Usageused":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total Java memory pool used."},

			"PeakUsageinit":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The initial peak Java memory pool allocated"},
			"PeakUsagemax":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum peak Java  memory pool available."},
			"PeakUsagecommitted": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total peak Java memory pool committed to be used"},
			"PeakUsageused":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total peak Java memory pool used."},
		},

		Tags: map[string]interface{}{
			"jolokia_agent_url": inputs.NewTagInfo("jolokia agent url path"),
			"name":              inputs.NewTagInfo("the name of space"),
		},
	}
}
