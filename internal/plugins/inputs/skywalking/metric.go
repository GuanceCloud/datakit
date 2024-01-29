// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package skywalking

import (
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	commonv3 "github.com/GuanceCloud/tracing-protos/skywalking-gen-go/common/v3"
	agentv3 "github.com/GuanceCloud/tracing-protos/skywalking-gen-go/language/agent/v3"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func meterDataToPoint(data *agentv3.MeterData) *point.Point {
	singel := data.GetSingleValue()
	if singel == nil {
		return nil
	}
	metric := &MetricMeasurement{
		name:   data.Service,
		tags:   make(map[string]string),
		fields: make(map[string]interface{}),
	}
	labels := singel.GetLabels()
	for _, label := range labels {
		metric.tags[label.GetName()] = label.GetValue()
	}
	ts := data.GetTimestamp()

	metric.fields[singel.Name] = singel.GetValue()
	metric.ts = time.UnixMilli(ts) // time

	return metric.Point()
}

// func processMetricsV3(jvm *agentv3.JVMMetricCollection, start time.Time) []inputs.Measurement {.
func processMetricsV3(jvm *agentv3.JVMMetricCollection, start time.Time, ipt *Input) []*point.Point {
	var metrics []*point.Point
	for _, jm := range jvm.Metrics {
		if jm.Cpu != nil {
			metrics = append(metrics, extractJVMCpuMetric(jvm.Service, start, jm.Cpu, ipt))
		}
		if len(jm.Memory) != 0 {
			metrics = append(metrics, extractJVMMemoryMetrics(jvm.Service, start, jm.Memory, ipt)...)
		}
		if len(jm.MemoryPool) != 0 {
			metrics = append(metrics, extractJVMMemoryPoolMetrics(jvm.Service, start, jm.MemoryPool, ipt)...)
		}
		if len(jm.Gc) != 0 {
			metrics = append(metrics, extractJVMGCMetrics(jvm.Service, start, jm.Gc, ipt)...)
		}
		if jm.Thread != nil {
			metrics = append(metrics, extractJVMThread(jvm.Service, start, jm.Thread, ipt))
		}
		if jm.Clazz != nil {
			metrics = append(metrics, extractJVMClass(jvm.Service, start, jm.Clazz, ipt))
		}
	}

	return metrics
}

type jvmMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
	ipt    *Input
}

// Point implement MeasurementV2.
func (m *jvmMeasurement) Point() *point.Point {
	opts := append(point.DefaultMetricOptions(), point.WithTime(m.ts), point.WithExtraTags(datakit.GlobalHostTags()))
	return point.NewPointV2(m.name, append(point.NewTags(m.tags), point.NewKVs(m.fields)...), opts...)
}

// nolint:lll
func (*jvmMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: jvmMetricName,
		Fields: map[string]interface{}{
			"cpu_usage_percent":               &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "The percentage of CPU usage by the JVM process."},
			"pid":                             &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "The process ID (PID) of the JVM."},
			"priority":                        &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "The priority of the JVM process."},
			"heap_max":                        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The maximum amount of heap memory that can be used by the JVM."},
			"heap_used":                       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The amount of heap memory currently being used."},
			"heap_committed":                  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The amount of heap memory that is currently committed for the JVM."},
			"heap_init":                       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The initial amount of heap memory allocated to the JVM upon start."},
			"stack_used":                      &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The amount of stack memory currently being used."},
			"stack_committed":                 &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The amount of stack memory that is currently committed for the JVM."},
			"stack_init":                      &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The initial amount of stack memory allocated to the JVM upon start."},
			"stack_max":                       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The maximum amount of stack memory that can be used by the JVM."},
			"pool_code_cache_usage_committed": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The amount of memory that is committed for the code cache area."},
			"pool_code_cache_usage_init":      &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The initial amount of memory allocated for the code cache area when the JVM started."},
			"pool_code_cache_usage_max":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The maximum amount of memory that can be used for the code cache area."},
			"pool_code_cache_usage_used":      &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The amount of memory currently used for the code cache area."},
			"pool_metaspace_usage_init":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The initial amount of memory allocated for the metaspace area when the JVM started."},
			"pool_metaspace_usage_committed":  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The amount of memory that is committed for the metaspace area."},
			"pool_metaspace_usage_max":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The maximum amount of memory that can be used for the metaspace area."},
			"pool_metaspace_usage_used":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The amount of memory currently used for the metaspace area."},
			"pool_permgen_usage_committed":    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The amount of memory that is committed for the permanent generation area."},
			"pool_permgen_usage_init":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The initial amount of memory allocated for the permanent generation area when the JVM started."},
			"pool_permgen_usage_used":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The amount of memory currently used for the permanent generation area."},
			"pool_permgen_usage_max":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The maximum amount of memory that can be used for the permanent generation area."},
			"pool_newgen_usage_committed":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The amount of memory that is committed for the young (new) generation area."},
			"pool_newgen_usage_used":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The amount of memory currently used for the young (new) generation area."},
			"pool_newgen_usage_max":           &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The maximum amount of memory that can be used for the young (new) generation area."},
			"pool_newgen_usage_init":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The initial amount of memory allocated for the young (new) generation area when the JVM started."},
			"pool_survivor_usage_max":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The maximum amount of memory that can be used for the survivor space area of the young generation."},
			"pool_survivor_usage_committed":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The amount of memory that is committed for the survivor space area of the young generation."},
			"pool_survivor_usage_used":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The amount of memory currently used for the survivor space area of the young generation."},
			"pool_survivor_usage_init":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The initial amount of memory allocated for the survivor space area of the young generation when the JVM started."},
			"pool_oldgen_usage_max":           &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The maximum amount of memory that can be used for the old generation area."},
			"pool_oldgen_usage_committed":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The amount of memory that is committed for the old generation area, which is used for objects that have survived multiple garbage collection cycles."},
			"pool_oldgen_usage_used":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The amount of memory currently being used by the old generation area."},
			"pool_oldgen_usage_init":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The initial amount of memory allocated for the old generation area when the JVM started."},
			"gc_new_count":                    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The count of young generation garbage collection events."},
			"gc_old_count":                    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The count of old generation garbage collection events."},
			"thread_blocked_state_count":      &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The number of threads that are currently in a blocked state."},
			"thread_runnable_state_count":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The number of threads that are currently in a runnable state."},
			"thread_peak_count":               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The peak number of live threads since the JVM started or peak was reset."},
			"thread_time_waiting_state_count": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The number of threads that are currently in a timed waiting state."},
			"thread_waiting_state_count":      &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The number of threads that are currently in a waiting state."},
			"thread_live_count":               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The current number of live threads, including both daemon and non-daemon threads."},
			"thread_daemon_count":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The total number of classes loaded since the JVM started."},
			"class_loaded_count":              &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The total number of classes loaded since the JVM started."},
			"class_total_unloaded_count":      &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The maximum amount of stack memory that can be used by the JVM."},
			"class_total_loaded_count":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The maximum amount of stack memory that can be used by the JVM."},
		},
		Tags: map[string]interface{}{
			"service": inputs.NewTagInfo("service name"),
		},
	}
}

func extractJVMCpuMetric(service string, start time.Time, cpu *commonv3.CPU, ipt *Input) *point.Point {
	metric := &jvmMeasurement{
		name:   jvmMetricName,
		tags:   map[string]string{"service": service},
		fields: map[string]interface{}{"cpu_usage_percent": cpu.UsagePercent},
		ts:     start,
		ipt:    ipt,
	}
	return metric.Point()
}

func extractJVMMemoryMetrics(service string, start time.Time, memory []*agentv3.Memory, ipt *Input) []*point.Point {
	isHeap := func(isHeap bool) string {
		if isHeap {
			return "heap"
		} else {
			return "stack"
		}
	}

	var m []*point.Point
	for _, v := range memory {
		if v == nil {
			continue
		}
		metric := &jvmMeasurement{
			name: jvmMetricName,
			tags: map[string]string{"service": service},
			fields: map[string]interface{}{
				isHeap(v.IsHeap) + "_init":      v.Init,
				isHeap(v.IsHeap) + "_max":       v.Max,
				isHeap(v.IsHeap) + "_used":      v.Used,
				isHeap(v.IsHeap) + "_committed": v.Committed,
			},
			ts:  start,
			ipt: ipt,
		}
		m = append(m, metric.Point())
	}

	return m
}

func extractJVMMemoryPoolMetrics(service string, start time.Time, memoryPool []*agentv3.MemoryPool, ipt *Input) []*point.Point {
	var m []*point.Point
	for _, v := range memoryPool {
		if v == nil {
			continue
		}
		metric := &jvmMeasurement{
			name: jvmMetricName,
			tags: map[string]string{"service": service},
			fields: map[string]interface{}{
				"pool_" + strings.ToLower(agentv3.PoolType_name[int32(v.Type)]) + "_init":      v.Init,
				"pool_" + strings.ToLower(agentv3.PoolType_name[int32(v.Type)]) + "_max":       v.Max,
				"pool_" + strings.ToLower(agentv3.PoolType_name[int32(v.Type)]) + "_used":      v.Used,
				"pool_" + strings.ToLower(agentv3.PoolType_name[int32(v.Type)]) + "_committed": v.Committed,
			},
			ts:  start,
			ipt: ipt,
		}
		m = append(m, metric.Point())
	}

	return m
}

func extractJVMGCMetrics(service string, start time.Time, gc []*agentv3.GC, ipt *Input) []*point.Point {
	var m []*point.Point
	for _, v := range gc {
		if v == nil {
			continue
		}
		metric := &jvmMeasurement{
			name: jvmMetricName,
			tags: map[string]string{"service": service},
			fields: map[string]interface{}{
				"gc_" + strings.ToLower(agentv3.GCPhase_name[int32(v.Phase)]) + "_count": v.Count,
			},
			ts:  start,
			ipt: ipt,
		}
		m = append(m, metric.Point())
	}

	return m
}

func extractJVMThread(service string, start time.Time, thread *agentv3.Thread, ipt *Input) *point.Point {
	metric := &jvmMeasurement{
		name: jvmMetricName,
		tags: map[string]string{"service": service},
		fields: map[string]interface{}{
			"thread_live_count":               thread.LiveCount,
			"thread_daemon_count":             thread.DaemonCount,
			"thread_peak_count":               thread.PeakCount,
			"thread_runnable_state_count":     thread.RunnableStateThreadCount,
			"thread_blocked_state_count":      thread.BlockedStateThreadCount,
			"thread_waiting_state_count":      thread.WaitingStateThreadCount,
			"thread_time_waiting_state_count": thread.TimedWaitingStateThreadCount,
		},
		ts:  start,
		ipt: ipt,
	}
	return metric.Point()
}

func extractJVMClass(service string, start time.Time, class *agentv3.Class, ipt *Input) *point.Point {
	metric := &jvmMeasurement{
		name: jvmMetricName,
		tags: map[string]string{"service": service},
		fields: map[string]interface{}{
			"class_loaded_count":         class.LoadedClassCount,
			"class_total_unloaded_count": class.TotalUnloadedClassCount,
			"class_total_loaded_count":   class.TotalLoadedClassCount,
		},
		ts:  start,
		ipt: ipt,
	}
	return metric.Point()
}

var _ inputs.Measurement = &MetricMeasurement{}

type MetricMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// Point implement MeasurementV2.
func (m *MetricMeasurement) Point() *point.Point {
	opts := append(point.DefaultMetricOptions(), point.WithTime(m.ts), point.WithExtraTags(datakit.GlobalHostTags()))

	return point.NewPointV2(m.name, append(point.NewTags(m.tags), point.NewKVs(m.fields)...), opts...)
}

func (*MetricMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: jvmMetricName,
		Desc: "jvm metrics collected by SkyWalking language agent.",
		Type: "metric",
		Tags: map[string]interface{}{"service": &inputs.TagInfo{Desc: "service name"}},
		Fields: map[string]interface{}{
			"cpu_usage_percent": &inputs.FieldInfo{
				Type:     inputs.Rate,
				DataType: inputs.Float,
				Unit:     inputs.Percent,
				Desc:     "cpu usage percentile",
			},
			"heap/stack_init": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "heap or stack initialized amount of memory.",
			},
			"heap/stack_max": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "heap or stack max amount of memory.",
			},
			"heap/stack_used": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "heap or stack used amount of memory.",
			},
			"heap/stack_committed": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "heap or stack committed amount of memory.",
			},
			"pool_*_init": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "initialized amount of memory in variety of pool(code_cache_usage,newgen_usage,oldgen_usage,survivor_usage,permgen_usage,metaspace_usage).", // nolint:lll
			},
			"pool_*_max": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "max amount of memory in variety of pool(code_cache_usage,newgen_usage,oldgen_usage,survivor_usage,permgen_usage,metaspace_usage).", // nolint:lll
			},
			"pool_*_used": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "used amount of memory in variety of pool(code_cache_usage,newgen_usage,oldgen_usage,survivor_usage,permgen_usage,metaspace_usage).", // nolint:lll
			},
			"pool_*_committed": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "committed amount of memory in variety of pool(code_cache_usage,newgen_usage,oldgen_usage,survivor_usage,permgen_usage,metaspace_usage).", // nolint:lll
			},
			"gc_phrase_old/new_count": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "gc old or new count.",
			},
			"thread_live_count": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "thread live count.",
			},
			"thread_daemon_count": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "thread daemon count.",
			},
			"thread_peak_count": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "thread peak count.",
			},
			"thread_runnable_state_count": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "runnable state thread count.",
			},
			"thread_blocked_state_count": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "blocked state thread count",
			},
			"thread_waiting_state_count": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "waiting state thread count.",
			},
			"thread_time_waiting_state_count": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "time waiting state thread count.",
			},
			"class_loaded_count": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "loaded class count.",
			},
			"class_total_unloaded_class_count": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "total unloaded class count.",
			},
			"class_total_loaded_count": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "total loaded class count.",
			},
		},
	}
}
