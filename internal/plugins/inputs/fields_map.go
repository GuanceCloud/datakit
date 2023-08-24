// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package inputs

////////////////////////////////////////////////////////////////////////////////

// AppendGeneralFields appends general fields. Assign only if not exist.
func AppendGeneralFields(in map[string]interface{}) {
	for _, v := range generalFieldsInstance {
		v.Append(in)
	}
}

// GetGeneralOptionalFieldsKeys returns general optional fields keys.
func GetGeneralOptionalFieldsKeys() []string {
	var out []string

	for _, v := range generalFieldsInstance {
		out = append(out, v.OptionalFieldsKeys()...)
	}

	return out
}

var generalFieldsInstance = []IFieldsMap{
	newGoRuntimeMemStatsFields(),
	newGoBaseStatsFields(),
	newProcessStatsFields(),
	newPromHTTPStatsFields(),
	newGRPCStatsFields(),
}

////////////////////////////////////////////////////////////////////////////////

type IFieldsMap interface {
	Append(in map[string]interface{})
	OptionalFieldsKeys() []string
}

func appendBuiltIn(in, builtin map[string]interface{}) {
	for k, v := range builtin {
		if _, ok := in[k]; !ok {
			// Assign Only if key not exist in origin.
			in[k] = v
		}
	}
}

func GetOptionalFieldsKeys(in map[string]interface{}) []string {
	var out []string

	for k, v := range in {
		out = append(out, k)

		ipFi, ok := v.(*FieldInfo)
		if ok {
			switch ipFi.Type {
			case Histogram:
				out = append(out, k+"_bucket")
				out = append(out, k+"_count")
				out = append(out, k+"_sum")
			case Summary:
				out = append(out, k+"_count")
				out = append(out, k+"_sum")
			}
		}
	}

	return out
}

////////////////////////////////////////////////////////////////////////////////

var _ IFieldsMap = (*goRuntimeMemStatsFields)(nil)

type goRuntimeMemStatsFields struct {
	builtin map[string]interface{}
}

func newGoRuntimeMemStatsFields() *goRuntimeMemStatsFields {
	//nolint:lll
	return &goRuntimeMemStatsFields{
		builtin: map[string]interface{}{
			"go_memstats_alloc_bytes_total":   &FieldInfo{DataType: Float, Type: Count, Unit: NCount, Desc: "Total number of bytes allocated, even if freed."},
			"go_memstats_alloc_bytes":         &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Number of bytes allocated and still in use."},
			"go_memstats_buck_hash_sys_bytes": &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Number of bytes used by the profiling bucket hash table."},
			"go_memstats_frees_total":         &FieldInfo{DataType: Float, Type: Count, Unit: NCount, Desc: "Total number of frees."},
			"go_memstats_gc_cpu_fraction":     &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "The fraction of this program's available CPU time used by the GC since the program started."},
			"go_memstats_gc_sys_bytes":        &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Number of bytes used for garbage collection system metadata."},
			"go_memstats_heap_alloc_bytes":    &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Number of heap bytes allocated and still in use."},
			"go_memstats_heap_idle_bytes":     &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Number of heap bytes waiting to be used."},
			"go_memstats_heap_inuse_bytes":    &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Number of heap bytes that are in use."},
			"go_memstats_heap_objects":        &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Number of allocated objects."},
			"go_memstats_heap_released_bytes": &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Number of heap bytes released to OS."},
			"go_memstats_heap_sys_bytes":      &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Number of heap bytes obtained from system."},
			"go_memstats_lookups_total":       &FieldInfo{DataType: Float, Type: Count, Unit: NCount, Desc: "Total number of pointer lookups."},
			"go_memstats_mallocs_total":       &FieldInfo{DataType: Float, Type: Count, Unit: NCount, Desc: "Total number of mallocs."},
			"go_memstats_mcache_inuse_bytes":  &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Number of bytes in use by mcache structures."},
			"go_memstats_mcache_sys_bytes":    &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Number of bytes used for mcache structures obtained from system."},
			"go_memstats_mspan_inuse_bytes":   &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Number of bytes in use by mspan structures."},
			"go_memstats_mspan_sys_bytes":     &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Number of bytes used for mspan structures obtained from system."},
			"go_memstats_next_gc_bytes":       &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Number of heap bytes when next garbage collection will take place."},
			"go_memstats_other_sys_bytes":     &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Number of bytes used for other system allocations."},
			"go_memstats_stack_inuse_bytes":   &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Number of bytes in use by the stack allocator."},
			"go_memstats_stack_sys_bytes":     &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Number of bytes obtained from system for stack allocator."},
			"go_memstats_sys_bytes":           &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Number of bytes obtained from system."},
		},
	}
}

func (fs *goRuntimeMemStatsFields) Append(in map[string]interface{}) {
	appendBuiltIn(in, fs.builtin)
}

func (fs *goRuntimeMemStatsFields) OptionalFieldsKeys() []string {
	return GetOptionalFieldsKeys(fs.builtin)
}

////////////////////////////////////////////////////////////////////////////////

var _ IFieldsMap = (*goBaseStatsFields)(nil)

type goBaseStatsFields struct {
	builtin map[string]interface{}
}

func newGoBaseStatsFields() *goBaseStatsFields {
	return &goBaseStatsFields{
		//nolint:lll
		builtin: map[string]interface{}{
			"go_gc_duration_seconds":           &FieldInfo{DataType: Float, Type: Summary, Unit: NCount, Desc: "A summary of the pause duration of garbage collection cycles."},
			"go_goroutines":                    &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Number of goroutines that currently exist."},
			"go_info":                          &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Information about the Go environment."},
			"go_memstats_last_gc_time_seconds": &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Number of seconds since 1970 of last garbage collection."},
			"go_threads":                       &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Number of OS threads created."},
		},
	}
}

func (fs *goBaseStatsFields) Append(in map[string]interface{}) {
	appendBuiltIn(in, fs.builtin)
}

func (fs *goBaseStatsFields) OptionalFieldsKeys() []string {
	return GetOptionalFieldsKeys(fs.builtin)
}

////////////////////////////////////////////////////////////////////////////////

var _ IFieldsMap = (*processStatsFields)(nil)

type processStatsFields struct {
	builtin map[string]interface{}
}

func newProcessStatsFields() *processStatsFields {
	return &processStatsFields{
		//nolint:lll
		builtin: map[string]interface{}{
			"process_cpu_seconds_total":        &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Total user and system CPU time spent in seconds"},
			"process_max_fds":                  &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Maximum number of open file descriptors"},
			"process_open_fds":                 &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Number of open file descriptors"},
			"process_resident_memory_bytes":    &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Resident memory size in bytes"},
			"process_start_time_seconds":       &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Start time of the process since unix epoch in seconds"},
			"process_virtual_memory_bytes":     &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Virtual memory size in bytes"},
			"process_virtual_memory_max_bytes": &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Maximum amount of virtual memory available in bytes"},
		},
	}
}

func (fs *processStatsFields) Append(in map[string]interface{}) {
	appendBuiltIn(in, fs.builtin)
}

func (fs *processStatsFields) OptionalFieldsKeys() []string {
	return GetOptionalFieldsKeys(fs.builtin)
}

////////////////////////////////////////////////////////////////////////////////

var _ IFieldsMap = (*promHTTPStatsFields)(nil)

type promHTTPStatsFields struct {
	builtin map[string]interface{}
}

func newPromHTTPStatsFields() *promHTTPStatsFields {
	//nolint:lll
	return &promHTTPStatsFields{
		builtin: map[string]interface{}{
			"promhttp_metric_handler_requests_in_flight": &FieldInfo{DataType: Float, Type: Gauge, Unit: NCount, Desc: "Current number of scrapes being served."},
			"promhttp_metric_handler_requests_total":     &FieldInfo{DataType: Float, Type: Count, Unit: NCount, Desc: "Total number of scrapes by HTTP status code."},
		},
	}
}

func (fs *promHTTPStatsFields) Append(in map[string]interface{}) {
	appendBuiltIn(in, fs.builtin)
}

func (fs *promHTTPStatsFields) OptionalFieldsKeys() []string {
	return GetOptionalFieldsKeys(fs.builtin)
}

////////////////////////////////////////////////////////////////////////////////

var _ IFieldsMap = (*grpcStatsFields)(nil)

type grpcStatsFields struct {
	builtin map[string]interface{}
}

func newGRPCStatsFields() *grpcStatsFields {
	//nolint:lll
	return &grpcStatsFields{
		builtin: map[string]interface{}{
			"grpc_server_handled_total":      &FieldInfo{DataType: Float, Type: Count, Unit: NCount, Desc: "Total number of RPCs completed on the server, regardless of success or failure."},
			"grpc_server_msg_received_total": &FieldInfo{DataType: Float, Type: Count, Unit: NCount, Desc: "Total number of RPC stream messages received on the server."},
			"grpc_server_msg_sent_total":     &FieldInfo{DataType: Float, Type: Count, Unit: NCount, Desc: "Total number of gRPC stream messages sent by the server."},
			"grpc_server_started_total":      &FieldInfo{DataType: Float, Type: Count, Unit: NCount, Desc: "Total number of RPCs started on the server."},
		},
	}
}

func (fs *grpcStatsFields) Append(in map[string]interface{}) {
	appendBuiltIn(in, fs.builtin)
}

func (fs *grpcStatsFields) OptionalFieldsKeys() []string {
	return GetOptionalFieldsKeys(fs.builtin)
}

////////////////////////////////////////////////////////////////////////////////
