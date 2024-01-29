// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jvm

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	javaRuntime          = "java_runtime"
	javaMemory           = "java_memory"
	javaGarbageCollector = "java_garbage_collector"
	javaThreading        = "java_threading"
	javaClassLoading     = "java_class_loading"
	javaMemoryPool       = "java_memory_pool"
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

type JavaThreadMemt struct {
	JvmMeasurement
}

type JavaClassLoadMemt struct {
	JvmMeasurement
}

type JavaMemoryPoolMemt struct {
	JvmMeasurement
}

////////////////////////////////////////////////////////////////////////////////

// Point implement MeasurementV2.
func (m *JvmMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (*JvmMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{}
}

////////////////////////////////////////////////////////////////////////////////

// Point implement MeasurementV2.
func (m *JavaRuntimeMemt) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (*JavaRuntimeMemt) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: javaRuntime,
		Fields: map[string]interface{}{
			"Uptime": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "The total runtime."},

			"CollectionUsageinit":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that the Java virtual machine initially requests from the operating system for memory management."},
			"CollectionUsagecommitted": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that is committed for the Java virtual machine to use."},
			"CollectionUsagemax":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum amount of memory in bytes that can be used for memory management."},
			"CollectionUsageused":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of used memory in bytes."},
		},
		Tags: map[string]interface{}{
			"jolokia_agent_url": inputs.TagInfo{Desc: "Jolokia agent url path."},
			"host":              inputs.TagInfo{Desc: "The hostname of the Jolokia agent/proxy running on."},
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

// Point implement MeasurementV2.
func (m *JavaMemoryMemt) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (*JavaMemoryMemt) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: javaMemory,
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

			"CollectionUsageinit":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that the Java virtual machine initially requests from the operating system for memory management."},
			"CollectionUsagecommitted": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that is committed for the Java virtual machine to use."},
			"CollectionUsagemax":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum amount of memory in bytes that can be used for memory management."},
			"CollectionUsageused":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of used memory in bytes."},
		},

		Tags: map[string]interface{}{
			"jolokia_agent_url": inputs.NewTagInfo("Jolokia agent url path."),
			"host":              inputs.NewTagInfo("The hostname of the Jolokia agent/proxy running on."),
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

// Point implement MeasurementV2.
func (m *JavaGcMemt) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (*JavaGcMemt) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: javaGarbageCollector,
		Fields: map[string]interface{}{
			"CollectionTime":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The approximate GC collection time elapsed."},
			"CollectionCount": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of GC that have occurred."},

			"CollectionUsageinit":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that the Java virtual machine initially requests from the operating system for memory management."},
			"CollectionUsagecommitted": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that is committed for the Java virtual machine to use."},
			"CollectionUsagemax":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum amount of memory in bytes that can be used for memory management."},
			"CollectionUsageused":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of used memory in bytes."},
		},

		Tags: map[string]interface{}{
			"jolokia_agent_url": inputs.NewTagInfo("Jolokia agent url path."),
			"name":              inputs.NewTagInfo("The name of GC generation."),
			"host":              inputs.NewTagInfo("The hostname of the Jolokia agent/proxy running on."),
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

// Point implement MeasurementV2.
func (m *JavaThreadMemt) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (*JavaThreadMemt) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: javaThreading,
		Fields: map[string]interface{}{
			"DaemonThreadCount":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The count of daemon thread."},
			"PeakThreadCount":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The peak count of thread."},
			"ThreadCount":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The count of thread."},
			"TotalStartedThreadCount": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total count of started thread."},

			"CollectionUsageinit":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that the Java virtual machine initially requests from the operating system for memory management."},
			"CollectionUsagecommitted": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that is committed for the Java virtual machine to use."},
			"CollectionUsagemax":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum amount of memory in bytes that can be used for memory management."},
			"CollectionUsageused":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of used memory in bytes."},
		},

		Tags: map[string]interface{}{
			"jolokia_agent_url": inputs.NewTagInfo("Jolokia agent url path."),
			"host":              inputs.NewTagInfo("The hostname of the Jolokia agent/proxy running on."),
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

// Point implement MeasurementV2.
func (m *JavaClassLoadMemt) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (*JavaClassLoadMemt) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: javaClassLoading,
		Fields: map[string]interface{}{
			"LoadedClassCount":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The count of loaded class."},
			"TotalLoadedClassCount": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total count of loaded class."},
			"UnloadedClassCount":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The count of unloaded class."},

			"CollectionUsageinit":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that the Java virtual machine initially requests from the operating system for memory management."},
			"CollectionUsagecommitted": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that is committed for the Java virtual machine to use."},
			"CollectionUsagemax":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum amount of memory in bytes that can be used for memory management."},
			"CollectionUsageused":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of used memory in bytes."},
		},

		Tags: map[string]interface{}{
			"jolokia_agent_url": inputs.NewTagInfo("Jolokia agent url path."),
			"host":              inputs.NewTagInfo("The hostname of the Jolokia agent/proxy running on."),
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

// Point implement MeasurementV2.
func (m *JavaMemoryPoolMemt) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (*JavaMemoryPoolMemt) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: javaMemoryPool,
		Fields: map[string]interface{}{
			"Usageinit":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The initial Java memory pool allocated."},
			"Usagemax":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum Java  memory pool available."},
			"Usagecommitted": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total Java memory pool committed to be used."},
			"Usageused":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total Java memory pool used."},

			"PeakUsageinit":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The initial peak Java memory pool allocated."},
			"PeakUsagemax":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum peak Java  memory pool available."},
			"PeakUsagecommitted": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total peak Java memory pool committed to be used."},
			"PeakUsageused":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total peak Java memory pool used."},

			"CollectionUsageinit":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that the Java virtual machine initially requests from the operating system for memory management."},
			"CollectionUsagecommitted": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that is committed for the Java virtual machine to use."},
			"CollectionUsagemax":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum amount of memory in bytes that can be used for memory management."},
			"CollectionUsageused":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of used memory in bytes."},
		},

		Tags: map[string]interface{}{
			"jolokia_agent_url": inputs.NewTagInfo("Jolokia agent url path."),
			"name":              inputs.NewTagInfo("The name of space."),
			"host":              inputs.NewTagInfo("The hostname of the Jolokia agent/proxy running on."),
		},
	}
}

////////////////////////////////////////////////////////////////////////////////
