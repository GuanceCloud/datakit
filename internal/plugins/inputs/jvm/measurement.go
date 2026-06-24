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
		Name: measurementJVM,
		Desc: "JVM runtime, memory, garbage collector, threading, " +
			"class loading, and memory-pool statistics collected through Jolokia Java platform MXBeans.",
		DescZh: "通过 Jolokia Java 平台 MXBean 采集的 JVM 运行时、内存、垃圾回收、线程、类加载和内存池指标。",
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
	tags["jolokia_agent_url"] = &inputs.TagInfo{Desc: "Jolokia agent URL used to collect the JVM metrics."}
	tags["host"] = &inputs.TagInfo{Desc: "Hostname reported by the Jolokia agent or proxy."}
	return tags
}

func (m *jvmMeasurement) getGCTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["name"] = &inputs.TagInfo{Desc: "Garbage collector or memory pool name associated with fields tagged by name."}
	return tags
}

func (m *jvmMeasurement) getPoolTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["name"] = &inputs.TagInfo{Desc: "Garbage collector or memory pool name associated with fields tagged by name."}
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
		Type:     inputs.Gauge,
		Unit:     inputs.DurationMS,
		Desc:     "Elapsed time since the JVM started.",
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
		Desc:     "Initial heap memory size requested by the JVM.",
	}
	fields["HeapMemoryUsageused"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Current heap memory used by the JVM.",
	}
	fields["HeapMemoryUsagemax"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Maximum heap memory available to the JVM, or -1 if undefined.",
	}
	fields["HeapMemoryUsagecommitted"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Heap memory currently committed for JVM use.",
	}

	fields["NonHeapMemoryUsageinit"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Initial non-heap memory size requested by the JVM.",
	}
	fields["NonHeapMemoryUsageused"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Current non-heap memory used by the JVM.",
	}
	fields["NonHeapMemoryUsagemax"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Maximum non-heap memory available to the JVM, or -1 if undefined.",
	}
	fields["NonHeapMemoryUsagecommitted"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Non-heap memory currently committed for JVM use.",
	}

	fields["ObjectPendingFinalizationCount"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Approximate current number of objects pending finalization.",
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
		Desc:     "Approximate accumulated elapsed time spent in garbage collection.",
	}
	fields["CollectionCount"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Total number of garbage collections that have occurred for the collector.",
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
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Current number of live daemon threads.",
	}
	fields["PeakThreadCount"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Peak live thread count since the JVM started or the peak was reset.",
	}
	fields["ThreadCount"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Current number of live threads.",
	}
	fields["TotalStartedThreadCount"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Total number of threads created and started since the JVM started.",
	}
	return fields
}

//nolint:lll
func (m *jvmMeasurement) getClassLoadingFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["LoadedClassCount"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Current number of classes loaded in the JVM.",
	}
	fields["TotalLoadedClassCount"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Total number of classes loaded since the JVM started.",
	}
	fields["UnloadedClassCount"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Total number of classes unloaded since the JVM started.",
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
		Desc:     "Initial memory size requested for this memory pool.",
	}
	fields["Usagemax"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Maximum memory available for this memory pool, or -1 if undefined.",
	}
	fields["Usagecommitted"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Memory currently committed for this memory pool.",
	}
	fields["Usageused"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Current memory used by this memory pool.",
	}

	fields["PeakUsageinit"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Initial memory size in the peak usage snapshot for this memory pool.",
	}
	fields["PeakUsagemax"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Maximum memory in the peak usage snapshot for this memory pool, or -1 if undefined.",
	}
	fields["PeakUsagecommitted"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Committed memory in the peak usage snapshot for this memory pool.",
	}
	fields["PeakUsageused"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Used memory in the peak usage snapshot for this memory pool.",
	}

	// Collection usage fields (shared across multiple measurements)
	fields["CollectionUsageinit"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Initial memory size in the post-GC collection usage snapshot for this memory pool.",
	}
	fields["CollectionUsagecommitted"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Committed memory in the post-GC collection usage snapshot for this memory pool.",
	}
	fields["CollectionUsagemax"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Maximum memory in the post-GC collection usage snapshot for this memory pool, or -1 if undefined.",
	}
	fields["CollectionUsageused"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Used memory in the post-GC collection usage snapshot for this memory pool.",
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
	return &inputs.MeasurementInfo{
		Desc:   "JVM metrics emitted with the runtime measurement name.",
		DescZh: "使用运行时指标集名称上报的 JVM 指标。",
	}
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
		Name:   javaRuntime,
		Cat:    point.Metric,
		Desc:   "Legacy JVM runtime measurement collected from `java.lang:type=Runtime` through Jolokia.",
		DescZh: "通过 Jolokia 从 `java.lang:type=Runtime` 采集的旧版 JVM 运行时指标。",
		Fields: map[string]interface{}{
			"Uptime": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Elapsed time since the JVM started."},

			"CollectionUsageinit":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that the Java virtual machine initially requests from the operating system for memory management."},
			"CollectionUsagecommitted": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that is committed for the Java virtual machine to use."},
			"CollectionUsagemax":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum amount of memory in bytes that can be used for memory management."},
			"CollectionUsageused":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of used memory in bytes."},
		},
		Tags: map[string]interface{}{
			"jolokia_agent_url": &inputs.TagInfo{Desc: "Jolokia agent URL used to collect the JVM metrics."},
			"host":              &inputs.TagInfo{Desc: "Hostname reported by the Jolokia agent or proxy."},
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
		Name:   javaMemory,
		Cat:    point.Metric,
		Desc:   "Legacy JVM heap and non-heap memory measurement collected from `java.lang:type=Memory` through Jolokia.",
		DescZh: "通过 Jolokia 从 `java.lang:type=Memory` 采集的旧版 JVM 堆内和非堆内存指标。",
		Fields: map[string]interface{}{
			"HeapMemoryUsageinit":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Initial heap memory size requested by the JVM."},
			"HeapMemoryUsageused":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Current heap memory used by the JVM."},
			"HeapMemoryUsagemax":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Maximum heap memory available to the JVM, or -1 if undefined."},
			"HeapMemoryUsagecommitted": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Heap memory currently committed for JVM use."},

			"NonHeapMemoryUsageinit":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Initial non-heap memory size requested by the JVM."},
			"NonHeapMemoryUsageused":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Current non-heap memory used by the JVM."},
			"NonHeapMemoryUsagemax":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Maximum non-heap memory available to the JVM, or -1 if undefined."},
			"NonHeapMemoryUsagecommitted": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Non-heap memory currently committed for JVM use."},

			"ObjectPendingFinalizationCount": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Approximate current number of objects pending finalization."},

			"CollectionUsageinit":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that the Java virtual machine initially requests from the operating system for memory management."},
			"CollectionUsagecommitted": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that is committed for the Java virtual machine to use."},
			"CollectionUsagemax":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum amount of memory in bytes that can be used for memory management."},
			"CollectionUsageused":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of used memory in bytes."},
		},

		Tags: map[string]interface{}{
			"jolokia_agent_url": inputs.NewTagInfo("Jolokia agent URL used to collect the JVM metrics."),
			"host":              inputs.NewTagInfo("Hostname reported by the Jolokia agent or proxy."),
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
		Name:   javaGarbageCollector,
		Cat:    point.Metric,
		Desc:   "Legacy JVM garbage-collector measurement collected from `java.lang:name=*,type=GarbageCollector` through Jolokia.",
		DescZh: "通过 Jolokia 从 `java.lang:name=*,type=GarbageCollector` 采集的旧版 JVM 垃圾回收指标。",
		Fields: map[string]interface{}{
			"CollectionTime":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "Approximate accumulated elapsed time spent in garbage collection."},
			"CollectionCount": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of garbage collections that have occurred for the collector."},

			"CollectionUsageinit":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that the Java virtual machine initially requests from the operating system for memory management."},
			"CollectionUsagecommitted": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that is committed for the Java virtual machine to use."},
			"CollectionUsagemax":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum amount of memory in bytes that can be used for memory management."},
			"CollectionUsageused":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of used memory in bytes."},
		},

		Tags: map[string]interface{}{
			"jolokia_agent_url": inputs.NewTagInfo("Jolokia agent URL used to collect the JVM metrics."),
			"name":              inputs.NewTagInfo("Garbage collector name."),
			"host":              inputs.NewTagInfo("Hostname reported by the Jolokia agent or proxy."),
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
		Name:   javaThreading,
		Cat:    point.Metric,
		Desc:   "Legacy JVM threading measurement collected from `java.lang:type=Threading` through Jolokia.",
		DescZh: "通过 Jolokia 从 `java.lang:type=Threading` 采集的旧版 JVM 线程指标。",
		Fields: map[string]interface{}{
			"DaemonThreadCount":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of live daemon threads."},
			"PeakThreadCount":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Peak live thread count since the JVM started or the peak was reset."},
			"ThreadCount":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of live threads."},
			"TotalStartedThreadCount": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of threads created and started since the JVM started."},

			"CollectionUsageinit":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that the Java virtual machine initially requests from the operating system for memory management."},
			"CollectionUsagecommitted": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that is committed for the Java virtual machine to use."},
			"CollectionUsagemax":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum amount of memory in bytes that can be used for memory management."},
			"CollectionUsageused":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of used memory in bytes."},
		},

		Tags: map[string]interface{}{
			"jolokia_agent_url": inputs.NewTagInfo("Jolokia agent URL used to collect the JVM metrics."),
			"host":              inputs.NewTagInfo("Hostname reported by the Jolokia agent or proxy."),
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
		Name:   javaClassLoading,
		Cat:    point.Metric,
		Desc:   "Legacy JVM class-loading measurement collected from `java.lang:type=ClassLoading` through Jolokia.",
		DescZh: "通过 Jolokia 从 `java.lang:type=ClassLoading` 采集的旧版 JVM 类加载指标。",
		Fields: map[string]interface{}{
			"LoadedClassCount":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of classes loaded in the JVM."},
			"TotalLoadedClassCount": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of classes loaded since the JVM started."},
			"UnloadedClassCount":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of classes unloaded since the JVM started."},

			"CollectionUsageinit":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that the Java virtual machine initially requests from the operating system for memory management."},
			"CollectionUsagecommitted": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that is committed for the Java virtual machine to use."},
			"CollectionUsagemax":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum amount of memory in bytes that can be used for memory management."},
			"CollectionUsageused":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of used memory in bytes."},
		},

		Tags: map[string]interface{}{
			"jolokia_agent_url": inputs.NewTagInfo("Jolokia agent URL used to collect the JVM metrics."),
			"host":              inputs.NewTagInfo("Hostname reported by the Jolokia agent or proxy."),
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
		Name:   javaMemoryPool,
		Cat:    point.Metric,
		Desc:   "Legacy JVM memory-pool measurement collected from `java.lang:name=*,type=MemoryPool` through Jolokia.",
		DescZh: "通过 Jolokia 从 `java.lang:name=*,type=MemoryPool` 采集的旧版 JVM 内存池指标。",
		Fields: map[string]interface{}{
			"Usageinit":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Initial memory size requested for this memory pool."},
			"Usagemax":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Maximum memory available for this memory pool, or -1 if undefined."},
			"Usagecommitted": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Memory currently committed for this memory pool."},
			"Usageused":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Current memory used by this memory pool."},

			"PeakUsageinit":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Initial memory size in the peak usage snapshot for this memory pool."},
			"PeakUsagemax":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Maximum memory in the peak usage snapshot for this memory pool, or -1 if undefined."},
			"PeakUsagecommitted": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Committed memory in the peak usage snapshot for this memory pool."},
			"PeakUsageused":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Used memory in the peak usage snapshot for this memory pool."},

			"CollectionUsageinit":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that the Java virtual machine initially requests from the operating system for memory management."},
			"CollectionUsagecommitted": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory in bytes that is committed for the Java virtual machine to use."},
			"CollectionUsagemax":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum amount of memory in bytes that can be used for memory management."},
			"CollectionUsageused":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of used memory in bytes."},
		},

		Tags: map[string]interface{}{
			"jolokia_agent_url": inputs.NewTagInfo("Jolokia agent URL used to collect the JVM metrics."),
			"name":              inputs.NewTagInfo("Memory pool name."),
			"host":              inputs.NewTagInfo("Hostname reported by the Jolokia agent or proxy."),
		},
	}
}

////////////////////////////////////////////////////////////////////////////////
