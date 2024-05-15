// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package statsd

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type (
	JVMMeasurement     struct{}
	JMXMeasurement     struct{}
	DDtraceMeasurement struct{}
)

// See also https://docs.datadoghq.com/opentelemetry/runtime_metrics/java/#runtime-metric-mappings

// Info ...
// nolint:lll
func (m *JVMMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "jvm",
		Fields: map[string]interface{}{
			"heap_memory":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total Java heap memory used."}, // down jvm & jmx mertrics
			"heap_memory_committed":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total Java heap memory committed to be used."},
			"heap_memory_init":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The initial Java heap memory allocated."},
			"heap_memory_max":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum Java heap memory available."},
			"non_heap_memory":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total Java non-heap memory used. Non-heap memory is: `Metaspace + CompressedClassSpace + CodeCache`."},
			"non_heap_memory_committed":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total Java non-heap memory committed to be used."},
			"non_heap_memory_init":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The initial Java non-heap memory allocated."},
			"non_heap_memory_max":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum Java non-heap memory available."},
			"gc_old_gen_size":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The ond gen size in garbage collection."},
			"gc_eden_size":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The 'eden' size in garbage collection."},
			"gc_survivor_size":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The survivor size in garbage collection."},
			"gc_metaspace_size":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The `metaspace` size in garbage collection."},
			"thread_count":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of live threads."},
			"loaded_classes":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of classes currently loaded."}, // up jvm & jmx mertrics
			"cpu_load_system":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Recent CPU utilization for the whole system."},
			"cpu_load_process":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Recent CPU utilization for the process."},
			"buffer_pool_direct_used":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Measure of memory used by direct buffers."},
			"buffer_pool_direct_count":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of direct buffers in the pool."},
			"buffer_pool_direct_capacity": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Measure of total memory capacity of direct buffers."},
			"buffer_pool_mapped_used":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Measure of memory used by mapped buffers."},
			"buffer_pool_mapped_count":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of mapped buffers in the pool."},
			"buffer_pool_mapped_capacity": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Measure of total memory capacity of mapped buffers."},
			"gc_parnew_time":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The approximate accumulated garbage collection time elapsed."},
			"gc_cms_count":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of garbage collections that have occurred."},
			"gc_major_collection_count":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The rate of major garbage collections. Set new_gc_metrics: true to receive this metric."},                          // jmx
			"gc_minor_collection_count":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The rate of minor garbage collections. Set new_gc_metrics: true to receive this metric."},                          // jmx
			"gc_major_collection_time":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.PartPerMillion, Desc: "The fraction of time spent in major garbage collection. Set new_gc_metrics: true to receive this metric."}, // jmx
			"gc_minor_collection_time":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.PartPerMillion, Desc: "The fraction of time spent in minor garbage collection. Set new_gc_metrics: true to receive this metric."}, // jmx
			"os_open_file_descriptors":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of file descriptors used by this process (only available for processes run as the dd-agent user)"},

			// Following metrics not found in official docs
			"daemon_code_cache_used": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of daemon threads."},
			"total_thread_count":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of total threads."},
			"peak_thread_count":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The peak number of live threads."},
			"gc_code_cache_used":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "GC code cache used."},
			"daemon_thread_count":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Daemon thread count."},
		},
		Tags: map[string]interface{}{
			"host":        inputs.TagInfo{Desc: "Host name."},
			"instance":    inputs.TagInfo{Desc: "Instance name."},
			"jmx_domain":  inputs.TagInfo{Desc: "JMX domain."},
			"metric_type": inputs.TagInfo{Desc: "Metric type."},
			"name":        inputs.TagInfo{Desc: "Type name."},
			"runtime-id":  inputs.TagInfo{Desc: "Runtime id."},
			"service":     inputs.TagInfo{Desc: "Service name."},
			"type":        inputs.TagInfo{Desc: "Object type."},
		},
	}
}

// See also https://docs.datadoghq.com/integrations/java/?tab=host#metrics

// Info ...
// nolint:lll
func (m *JMXMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "jmx",
		Fields: map[string]interface{}{
			// buffer_pool_direct_capacity 这个没找到解释
			"heap_memory":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total Java heap memory used."},
			"heap_memory_committed":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total Java heap memory committed to be used."},
			"heap_memory_init":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The initial Java heap memory allocated."},
			"heap_memory_max":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum Java heap memory available."},
			"non_heap_memory":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total Java non-heap memory used. Non-heap memory is calculated as follows: 'Metaspace' + CompressedClassSpace + CodeCache"},
			"non_heap_memory_committed": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total Java non-heap memory committed to be used."},
			"non_heap_memory_init":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The initial Java non-heap memory allocated."},
			"non_heap_memory_max":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum Java non-heap memory available."},
			"thread_count":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of live threads."},
			"gc_cms.count":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of garbage collections that have occurred."},
			"gc_major_collection_count": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The rate of major garbage collections. Set new_gc_metrics: true to receive this metric."},
			"gc_minor_collection_count": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The rate of minor garbage collections. Set new_gc_metrics: true to receive this metric."},
			"gc_parnew.time":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The approximate accumulated garbage collection time elapsed."},
			"gc_major_collection_time":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.PartPerMillion, Desc: "The fraction of time spent in major garbage collection. Set new_gc_metrics: true to receive this metric."},
			"gc_minor_collection_time":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.PartPerMillion, Desc: "The fraction of time spent in minor garbage collection. Set new_gc_metrics: true to receive this metric."},
		},
		Tags: map[string]interface{}{
			"host":        inputs.TagInfo{Desc: "Host name."},
			"instance":    inputs.TagInfo{Desc: "Instance name."},
			"jmx_domain":  inputs.TagInfo{Desc: "JMX domain."},
			"metric_type": inputs.TagInfo{Desc: "Metric type."},
			"name":        inputs.TagInfo{Desc: "Type name."},
			"runtime-id":  inputs.TagInfo{Desc: "Runtime id."},
			"service":     inputs.TagInfo{Desc: "Service name."},
			"type":        inputs.TagInfo{Desc: "Object type."},
		},
	}
}

// See also https://docs.datadoghq.com/opentelemetry/runtime_metrics/java/
// See also https://docs.datadoghq.com/developers/metrics/types/?tab=count#metric-types

// Info ...
// nolint:lll
func (m *DDtraceMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "ddtrace",
		Fields: map[string]interface{}{
			"tracer_queue_enqueued_spans":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Tracer queue enqueued spans."},
			"tracer_queue_enqueued_traces":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Tracer queue enqueued traces."},
			"tracer_scope_activate_count":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Tracer scope activate count."},
			"tracer_scope_close_count":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Tracer scope close count."},
			"tracer_span_pending_created":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Tracer span pending created."},
			"tracer_span_pending_finished":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Tracer span pending finished."},
			"tracer_trace_pending_created":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Tracer trace pending created."},
			"tracer_agent_discovery_time":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Tracer agent discovery time."},
			"tracer_trace_agent_discovery_time":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Tracer trace agent discovery time."},
			"tracer_tracer_trace_buffer_fill_time": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Tracer trace buffer fill time."},
			"tracer_trace_agent_send_time":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Tracer trace agent send time."},
			"tracer_api_errors_total":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Tracer api errors total."},
			"tracer_flush_bytes_total":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Tracer flush bytes total."},
			"tracer_flush_traces_total":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Tracer flush traces total."},
			"tracer_api_requests_total":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Tracer api requests total."},
			"tracer_queue_enqueued_bytes":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Tracer queue enqueued bytes."},
			"tracer_queue_max_length":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Tracer queue max length."},
		},
		Tags: map[string]interface{}{
			"host":                    inputs.TagInfo{Desc: "Host name."},
			"lang_interpreter":        inputs.TagInfo{Desc: "Lang interpreter."},
			"service":                 inputs.TagInfo{Desc: "Service name."},
			"tracer_version":          inputs.TagInfo{Desc: "Tracer version."},
			"lang_version":            inputs.TagInfo{Desc: "Lang version."},
			"metric_type":             inputs.TagInfo{Desc: "Metric type."},
			"stat":                    inputs.TagInfo{Desc: "Stat."},
			"priority":                inputs.TagInfo{Desc: "Priority."},
			"lang_interpreter_vendor": inputs.TagInfo{Desc: "Lang interpreter vendor."},
			"lang":                    inputs.TagInfo{Desc: "Lang type."},
			"endpoint":                inputs.TagInfo{Desc: "Endpoint."},
		},
	}
}

func MergeAllMeasurementInfo(info *inputs.MeasurementInfo) *inputs.MeasurementInfo {
	return MergeMeasurementInfo(info, (&JVMMeasurement{}).Info(), (&JMXMeasurement{}).Info(), (&DDtraceMeasurement{}).Info())
}

func MergeMeasurementInfo(infos ...*inputs.MeasurementInfo) *inputs.MeasurementInfo {
	if len(infos) == 0 {
		return &inputs.MeasurementInfo{}
	}

	retInfo := infos[len(infos)-1]
	for i := len(infos) - 2; i >= 0; i-- {
		for k, v := range infos[i].Fields {
			retInfo.Fields[k] = v
		}
		for k, v := range infos[i].Tags {
			retInfo.Tags[k] = v
		}
	}
	retInfo.Name = infos[0].Name

	return retInfo
}

func MergeSlice(a, b []string) []string {
	m := make(map[string]bool)
	for _, v := range a {
		m[v] = true
	}
	for _, v := range b {
		m[v] = true
	}
	c := make([]string, 0)
	for k := range m {
		c = append(c, k)
	}
	return c
}

func GetJVMOptionalFields() []string {
	m := &JVMMeasurement{}
	info := m.Info()
	_ = info
	s := make([]string, 0)
	for k := range m.Info().Fields {
		s = append(s, k)
	}
	return s
}

func GetJVMOptionalTags() []string {
	m := &JVMMeasurement{}
	s := make([]string, 0)
	for k := range m.Info().Tags {
		s = append(s, k)
	}
	return s
}

func GetJMXOptionalFields() []string {
	m := &JMXMeasurement{}
	info := m.Info()
	_ = info
	s := make([]string, 0)
	for k := range m.Info().Fields {
		s = append(s, k)
	}
	return s
}

func GetJMXOptionalTags() []string {
	m := &JMXMeasurement{}
	s := make([]string, 0)
	for k := range m.Info().Tags {
		s = append(s, k)
	}
	return s
}

func GetDDtraceOptionalFields() []string {
	m := &DDtraceMeasurement{}
	info := m.Info()
	_ = info
	s := make([]string, 0)
	for k := range m.Info().Fields {
		s = append(s, k)
	}
	return s
}

func GetDDtraceOptionalTags() []string {
	m := &DDtraceMeasurement{}
	s := make([]string, 0)
	for k := range m.Info().Tags {
		s = append(s, k)
	}
	return s
}
