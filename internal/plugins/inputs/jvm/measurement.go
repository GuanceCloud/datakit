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

	measurementJVM = "jvm"

	TagGroupGC   = "gc"
	TagGroupPool = "pool"
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

type jvmMeasurement struct{}

////////////////////////////////////////////////////////////////////////////////

// Info returns the unified JVM measurement info (v2).
func (m *jvmMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   measurementJVM,
		Desc:   "Metric set including JVM runtime, memory, garbage collector, threading, class loading, and memory pool statistics, unified in v2",
		DescZh: "指标集包含 JVM runtime、memory、garbage collector、threading、class loading 和 memory pool 相关指标，v2 版本统一",
		Cat:    point.Metric,
		Tags:   m.getTags(),
		Fields: m.getFields(),
	}
}

func (m *jvmMeasurement) getTags() map[string]interface{} {
	return mergeMaps(
		m.getCommonTags(),
		m.getGCTags(),
		m.getPoolTags(),
	)
}

func (m *jvmMeasurement) getCommonTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["jolokia_agent_url"] = &inputs.TagInfo{Desc: "Jolokia agent url path."}
	tags["host"] = &inputs.TagInfo{Desc: "The hostname of the Jolokia agent/proxy running on."}
	return tags
}

func (m *jvmMeasurement) getGCTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["name"] = &inputs.TagInfo{Desc: "The name of GC generation."}
	return tags
}

func (m *jvmMeasurement) getPoolTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["name"] = &inputs.TagInfo{Desc: "The name of memory pool."}
	return tags
}

func mergeMaps(fieldMaps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range fieldMaps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

func (m *jvmMeasurement) getFields() map[string]interface{} {
	return mergeMaps(
		m.getRuntimeFields(),
		m.getMemoryFields(),
		m.getGCFields(),
		m.getThreadingFields(),
		m.getClassLoadingFields(),
		m.getMemoryPoolFields(),
	)
}

func (m *jvmMeasurement) getRuntimeFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["Uptime"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.DurationMS,
		Desc:     "The total runtime.",
	}
	return fields
}

//nolint:lll
func (m *jvmMeasurement) getMemoryFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["HeapMemoryUsageinit"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The initial Java heap memory allocated.",
	}
	fields["HeapMemoryUsageused"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The total Java heap memory used.",
	}
	fields["HeapMemoryUsagemax"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The maximum Java heap memory available.",
	}
	fields["HeapMemoryUsagecommitted"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The total Java heap memory committed to be used.",
	}

	fields["NonHeapMemoryUsageinit"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The initial Java non-heap memory allocated.",
	}
	fields["NonHeapMemoryUsageused"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The total Java non-heap memory used.",
	}
	fields["NonHeapMemoryUsagemax"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The maximum Java non-heap memory available.",
	}
	fields["NonHeapMemoryUsagecommitted"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The total Java non-heap memory committed to be used.",
	}

	fields["ObjectPendingFinalizationCount"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The count of object pending finalization.",
	}
	return fields
}

//nolint:lll
func (m *jvmMeasurement) getGCFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["CollectionTime"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.DurationMS,
		Desc:     "The approximate GC collection time elapsed.",
	}
	fields["CollectionCount"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of GC that have occurred.",
	}

	// GC fields are tagged by name (GC generation name)
	m.addTaggedbyToFields(fields, TagGroupGC)
	return fields
}

//nolint:lll
func (m *jvmMeasurement) getThreadingFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["DaemonThreadCount"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The count of daemon thread.",
	}
	fields["PeakThreadCount"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The peak count of thread.",
	}
	fields["ThreadCount"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The count of thread.",
	}
	fields["TotalStartedThreadCount"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The total count of started thread.",
	}
	return fields
}

//nolint:lll
func (m *jvmMeasurement) getClassLoadingFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["LoadedClassCount"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The count of loaded class.",
	}
	fields["TotalLoadedClassCount"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The total count of loaded class.",
	}
	fields["UnloadedClassCount"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The count of unloaded class.",
	}
	return fields
}

//nolint:lll
func (m *jvmMeasurement) getMemoryPoolFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["Usageinit"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The initial Java memory pool allocated.",
	}
	fields["Usagemax"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The maximum Java memory pool available.",
	}
	fields["Usagecommitted"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The total Java memory pool committed to be used.",
	}
	fields["Usageused"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The total Java memory pool used.",
	}

	fields["PeakUsageinit"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The initial peak Java memory pool allocated.",
	}
	fields["PeakUsagemax"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The maximum peak Java memory pool available.",
	}
	fields["PeakUsagecommitted"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The total peak Java memory pool committed to be used.",
	}
	fields["PeakUsageused"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The total peak Java memory pool used.",
	}

	// Collection usage fields (shared across multiple measurements)
	fields["CollectionUsageinit"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The amount of memory in bytes that the Java virtual machine initially requests from the operating system for memory management.",
	}
	fields["CollectionUsagecommitted"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The amount of memory in bytes that is committed for the Java virtual machine to use.",
	}
	fields["CollectionUsagemax"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The maximum amount of memory in bytes that can be used for memory management.",
	}
	fields["CollectionUsageused"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The amount of used memory in bytes.",
	}

	m.addTaggedbyToFields(fields, TagGroupPool)
	return fields
}

func (m *jvmMeasurement) addTaggedbyToFields(fields map[string]interface{}, tagGroup string) {
	var tags map[string]interface{}

	// Only add TaggedBy for non-common tags (like name)
	switch tagGroup {
	case TagGroupGC:
		// GC fields are tagged by name (GC generation name)
		tags = m.getGCTags()
	case TagGroupPool:
		// Pool fields are tagged by name (memory pool name)
		tags = m.getPoolTags()
	default:
		return
	}

	// Extract tag keys
	taggedBy := make([]string, 0, len(tags))
	for tag := range tags {
		taggedBy = append(taggedBy, tag)
	}

	// Add Taggedby to each field
	for _, field := range fields {
		if fieldInfo, ok := field.(*inputs.FieldInfo); ok {
			fieldInfo.Taggedby = taggedBy
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

// Point implement MeasurementV2.
func (m *JvmMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPoint(m.name,
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

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (*JavaRuntimeMemt) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: javaRuntime,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"Uptime": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "The total runtime."},

			"CollectionUsageinit":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that the Java virtual machine initially requests from the operating system for memory management."},
			"CollectionUsagecommitted": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that is committed for the Java virtual machine to use."},
			"CollectionUsagemax":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum amount of memory in bytes that can be used for memory management."},
			"CollectionUsageused":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of used memory in bytes."},
		},
		Tags: map[string]interface{}{
			"jolokia_agent_url": &inputs.TagInfo{Desc: "Jolokia agent url path."},
			"host":              &inputs.TagInfo{Desc: "The hostname of the Jolokia agent/proxy running on."},
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

// Point implement MeasurementV2.
func (m *JavaMemoryMemt) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (*JavaMemoryMemt) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: javaMemory,
		Cat:  point.Metric,
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

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (*JavaGcMemt) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: javaGarbageCollector,
		Cat:  point.Metric,
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

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (*JavaThreadMemt) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: javaThreading,
		Cat:  point.Metric,
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

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (*JavaClassLoadMemt) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: javaClassLoading,
		Cat:  point.Metric,
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

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (*JavaMemoryPoolMemt) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: javaMemoryPool,
		Cat:  point.Metric,
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
