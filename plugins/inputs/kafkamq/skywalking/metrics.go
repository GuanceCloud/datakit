// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package skywalking handle SkyWalking tracing metrics.
package skywalking

import (
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	commonv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/common/v3"
	agentv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/language/agent/v3"
)

const jvmMetricName = "skywalking_jvm"

func processMetrics(jvm *agentv3.JVMMetricCollection) {
	var (
		m     []inputs.Measurement
		start = time.Now()
	)
	for _, jm := range jvm.Metrics {
		if jm.Cpu != nil {
			m = append(m, extractJVMCpuMetric(jvm.Service, start, jm.Cpu))
		}
		if len(jm.Memory) != 0 {
			m = append(m, extractJVMMemoryMetrics(jvm.Service, start, jm.Memory)...)
		}
		if len(jm.MemoryPool) != 0 {
			m = append(m, extractJVMMemoryPoolMetrics(jvm.Service, start, jm.MemoryPool)...)
		}
		if len(jm.Gc) != 0 {
			m = append(m, extractJVMGCMetrics(jvm.Service, start, jm.Gc)...)
		}
		if jm.Thread != nil {
			m = append(m, extractJVMThread(jvm.Service, start, jm.Thread))
		}
		if jm.Clazz != nil {
			m = append(m, extractJVMClass(jvm.Service, start, jm.Clazz))
		}
	}

	if len(m) != 0 {
		if err := inputs.FeedMeasurement(jvmMetricName, datakit.Metric, m, &dkio.Option{CollectCost: time.Since(start)}); err != nil {
			dkio.FeedLastError(jvmMetricName, err.Error())
		}
	}
}

type jvmMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *jvmMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOpt())
}

func (*jvmMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{}
}

func extractJVMCpuMetric(service string, start time.Time, cpu *commonv3.CPU) inputs.Measurement {
	return &jvmMeasurement{
		name:   jvmMetricName,
		tags:   map[string]string{"service": service},
		fields: map[string]interface{}{"cpu_usage_percent": cpu.UsagePercent},
		ts:     start,
	}
}

func extractJVMMemoryMetrics(service string, start time.Time, memory []*agentv3.Memory) []inputs.Measurement {
	isHeap := func(isHeap bool) string {
		if isHeap {
			return "heap"
		} else {
			return "stack"
		}
	}

	var m []inputs.Measurement
	for _, v := range memory {
		if v == nil {
			continue
		}
		m = append(m, &jvmMeasurement{
			name: jvmMetricName,
			tags: map[string]string{"service": service},
			fields: map[string]interface{}{
				isHeap(v.IsHeap) + "_init":      v.Init,
				isHeap(v.IsHeap) + "_max":       v.Max,
				isHeap(v.IsHeap) + "_used":      v.Used,
				isHeap(v.IsHeap) + "_committed": v.Committed,
			},
			ts: start,
		})
	}

	return m
}

func extractJVMMemoryPoolMetrics(service string, start time.Time, memoryPool []*agentv3.MemoryPool) []inputs.Measurement {
	var m []inputs.Measurement
	for _, v := range memoryPool {
		if v == nil {
			continue
		}
		m = append(m, &jvmMeasurement{
			name: jvmMetricName,
			tags: map[string]string{"service": service},
			fields: map[string]interface{}{
				"pool_" + strings.ToLower(agentv3.PoolType_name[int32(v.Type)]) + "_init":      v.Init,
				"pool_" + strings.ToLower(agentv3.PoolType_name[int32(v.Type)]) + "_max":       v.Max,
				"pool_" + strings.ToLower(agentv3.PoolType_name[int32(v.Type)]) + "_used":      v.Used,
				"pool_" + strings.ToLower(agentv3.PoolType_name[int32(v.Type)]) + "_committed": v.Committed,
			},
			ts: start,
		})
	}

	return m
}

func extractJVMGCMetrics(service string, start time.Time, gc []*agentv3.GC) []inputs.Measurement {
	var m []inputs.Measurement
	for _, v := range gc {
		if v == nil {
			continue
		}
		m = append(m, &jvmMeasurement{
			name: jvmMetricName,
			tags: map[string]string{"service": service},
			fields: map[string]interface{}{
				"gc_" + strings.ToLower(agentv3.GCPhase_name[int32(v.Phase)]) + "_count": v.Count,
			},
			ts: start,
		})
	}

	return m
}

func extractJVMThread(service string, start time.Time, thread *agentv3.Thread) inputs.Measurement {
	return &jvmMeasurement{
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
		ts: start,
	}
}

func extractJVMClass(service string, start time.Time, class *agentv3.Class) inputs.Measurement {
	return &jvmMeasurement{
		name: jvmMetricName,
		tags: map[string]string{"service": service},
		fields: map[string]interface{}{
			"class_loaded_count":         class.LoadedClassCount,
			"class_total_unloaded_count": class.TotalUnloadedClassCount,
			"class_total_loaded_count":   class.TotalLoadedClassCount,
		},
		ts: start,
	}
}

var _ inputs.Measurement = &MetricMeasurement{}

type MetricMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *MetricMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOpt())
}

func (*MetricMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: jvmMetricName,
		Desc: "jvm metrics collected by skywalking language agent.",
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
