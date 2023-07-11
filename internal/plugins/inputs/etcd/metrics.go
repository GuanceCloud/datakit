// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package etcd

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type etcdMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
	// ipt    *Input
}

// Point implement MeasurementV2.
func (m *etcdMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2([]byte(m.name),
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (m *etcdMeasurement) LineProto() (*dkpt.Point, error) {
	// return point.NewPoint(m.name, m.tags, m.fields, point.MOptElectionV2(m.election))
	return nil, fmt.Errorf("not implement")
}

//nolint:lll
func (m *etcdMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Desc: "etcd metrics.",
		Fields: map[string]interface{}{
			"etcd_cluster_version":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd cluster version."},
			"etcd_debugging_auth_revision":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging auth revision."},
			"etcd_debugging_lease_granted_total":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging lease granted total."},
			"etcd_debugging_lease_renewed_total":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging lease renewed total."},
			"etcd_debugging_lease_revoked_total":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging lease revoked total."},
			"etcd_debugging_mvcc_compact_revision":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging mvcc compact revision."},
			"etcd_debugging_mvcc_current_revision":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging mvcc current revision."},
			"etcd_debugging_mvcc_db_compaction_keys_total":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of db keys compacted."},
			"etcd_debugging_mvcc_db_compaction_last":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging mvcc db compaction last."},
			"etcd_debugging_mvcc_db_total_size_in_bytes":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total size of the underlying database in bytes."},
			"etcd_debugging_mvcc_delete_total":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of deletes seen by this member."},
			"etcd_debugging_mvcc_events_total":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of events sent by this member."},
			"etcd_debugging_mvcc_keys_total":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of keys."},
			"etcd_debugging_mvcc_pending_events_total":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of pending events to be sent."},
			"etcd_debugging_mvcc_put_total":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of puts seen by this member."},
			"etcd_debugging_mvcc_range_total":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of ranges seen by this member."},
			"etcd_debugging_mvcc_slow_watcher_total":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of unsynced slow watchers."},
			"etcd_debugging_mvcc_total_put_size_in_bytes":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging mvcc total put size in bytes."},
			"etcd_debugging_mvcc_txn_total":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of txns seen by this member."},
			"etcd_debugging_mvcc_watch_stream_total":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of watch streams."},
			"etcd_debugging_mvcc_watcher_total":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of watchers."},
			"etcd_debugging_server_lease_expired_total":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of expired leases."},
			"etcd_debugging_store_expires_total":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of expired keys."},
			"etcd_debugging_store_reads_total":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of reads action by (get/getRecursive), local to this member."},
			"etcd_debugging_store_watch_requests_total":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of incoming watch requests (new or reestablished)."},
			"etcd_debugging_store_watchers":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of currently active watchers."},
			"etcd_debugging_store_writes_total":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of writes (e.g. set/compareAndDelete) seen by this member."},
			"etcd_disk_defrag_inflight":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd disk defrag inflight"},
			"etcd_disk_wal_write_bytes_total":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of bytes written in WAL."},
			"etcd_grpc_proxy_cache_hits_total":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of cache hits."},
			"etcd_grpc_proxy_cache_keys_total":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of keys/ranges cached."},
			"etcd_grpc_proxy_cache_misses_total":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of cache misses."},
			"etcd_grpc_proxy_events_coalescing_total":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of events coalescing."},
			"etcd_grpc_proxy_watchers_coalescing_total":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of current watchers coalescing."},
			"etcd_mvcc_db_open_read_transactions":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd mvcc db open read transactions."},
			"etcd_mvcc_db_total_size_in_bytes":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd mvcc db total size in bytes."},
			"etcd_mvcc_db_total_size_in_use_in_bytes":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total size of the underlying database logically in use."},
			"etcd_mvcc_delete_total":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd mvcc delete total."},
			"etcd_mvcc_put_total":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd mvcc put total."},
			"etcd_mvcc_range_total":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd mvcc range total."},
			"etcd_mvcc_txn_total":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd mvcc txn total."},
			"etcd_network_client_grpc_received_bytes_total": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of bytes received from grpc clients."},
			"etcd_network_client_grpc_sent_bytes_total":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of bytes sent to grpc clients."},
			"etcd_server_go_version":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Which Go version server is running with. 1 with label with current version."},
			"etcd_server_has_leader":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Whether or not a leader exists. 1 is existence, 0 is not."},
			"etcd_server_health_failures":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of failed health checks."},
			"etcd_server_health_success":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of successful health checks."},
			"etcd_server_heartbeat_send_failures_total":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of leader heartbeat send failures (likely overloaded from slow disk)."},
			"etcd_server_id":                                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server id."},
			"etcd_server_is_leader":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Whether or not this member is a leader. 1 if is, 0 otherwise."},
			"etcd_server_is_learner":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server is learner."},
			"etcd_server_leader_changes_seen_total":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of leader changes seen."},
			"etcd_server_learner_promote_successes":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server learner promote successes."},
			"etcd_server_proposals_applied_total":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of consensus proposals applied."},
			"etcd_server_proposals_committed_total":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of consensus proposals committed."},
			"etcd_server_proposals_failed_total":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of failed proposals seen."},
			"etcd_server_proposals_pending":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The current number of pending proposals to commit."},
			"etcd_server_quota_backend_bytes":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Current backend storage quota size in bytes"},
			"etcd_server_read_indexes_failed_total":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of failed read indexes seen."},
			"etcd_server_slow_apply_total":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of slow apply requests (likely overloaded from slow disk)."},
			"etcd_server_slow_read_indexes_total":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server slow read indexes total."},
			"etcd_server_snapshot_apply_in_progress_total":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server snapshot apply in progress total."},
			"etcd_server_version":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Which version is running. 1 for 'server_version' label with current version."},
			"go_goroutines":                                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of goroutines that currently exist."},
			"go_info":                                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Information about the Go environment."},
			"go_memstats_alloc_bytes_total":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of bytes allocated, even if freed."},
			"go_memstats_alloc_bytes":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of bytes allocated and still in use."},
			"go_memstats_buck_hash_sys_bytes":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of bytes used by the profiling bucket hash table."},
			"go_memstats_frees_total":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of frees."},
			"go_memstats_gc_cpu_fraction":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The fraction of this program's available CPU time used by the GC since the program started."},
			"go_memstats_gc_sys_bytes":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of bytes used for garbage collection system metadata."},
			"go_memstats_heap_alloc_bytes":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of heap bytes allocated and still in use."},
			"go_memstats_heap_idle_bytes":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of heap bytes waiting to be used."},
			"go_memstats_heap_inuse_bytes":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of heap bytes that are in use."},
			"go_memstats_heap_objects":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of allocated objects."},
			"go_memstats_heap_released_bytes":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of heap bytes released to OS."},
			"go_memstats_heap_sys_bytes":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of heap bytes obtained from system."},
			"go_memstats_last_gc_time_seconds":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of seconds since 1970 of last garbage collection."},
			"go_memstats_lookups_total":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of pointer lookups."},
			"go_memstats_mallocs_total":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of mallocs."},
			"go_memstats_mcache_inuse_bytes":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of bytes in use by mcache structures."},
			"go_memstats_mcache_sys_bytes":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of bytes used for mcache structures obtained from system."},
			"go_memstats_mspan_inuse_bytes":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of bytes in use by mspan structures."},
			"go_memstats_mspan_sys_bytes":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of bytes used for mspan structures obtained from system."},
			"go_memstats_next_gc_bytes":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of heap bytes when next garbage collection will take place."},
			"go_memstats_other_sys_bytes":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of bytes used for other system allocations."},
			"go_memstats_stack_inuse_bytes":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of bytes in use by the stack allocator."},
			"go_memstats_stack_sys_bytes":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of bytes obtained from system for stack allocator."},
			"go_memstats_sys_bytes":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of bytes obtained from system."},
			"go_threads":                                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of OS threads created."},
			"grpc_server_handled_total":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of RPCs completed on the server, regardless of success or failure."},
			"grpc_server_msg_received_total":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of RPC stream messages received on the server."},
			"grpc_server_msg_sent_total":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of gRPC stream messages sent by the server."},
			"grpc_server_started_total":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of RPCs started on the server."},
			"os_fd_limit":                                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The file descriptor limit."},
			"os_fd_used":                                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of used file descriptors."},
			"process_cpu_seconds_total":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total user and system CPU time spent in seconds."},
			"process_max_fds":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Maximum number of open file descriptors."},
			"process_open_fds":                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of open file descriptors."},
			"process_resident_memory_bytes":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Resident memory size in bytes."},
			"process_start_time_seconds":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Start time of the process since unix epoch in seconds."},
			"process_virtual_memory_bytes":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Virtual memory size in bytes."},
			"process_virtual_memory_max_bytes":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Process virtual memory max bytes."},
			"promhttp_metric_handler_requests_in_flight":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Promhttp metric handler requests in flight."},
			"promhttp_metric_handler_requests_total":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Promhttp metric handler requests total."},
		},
		Tags: map[string]interface{}{
			"action":            inputs.NewTagInfo("Action."),
			"cluster_version":   inputs.NewTagInfo("Cluster version."),
			"code":              inputs.NewTagInfo("Code."),
			"grpc_code":         inputs.NewTagInfo("GRPC code."),
			"grpc_method":       inputs.NewTagInfo("GRPC method."),
			"grpc_service":      inputs.NewTagInfo("GRPC service name."),
			"grpc_type":         inputs.NewTagInfo("GRPC type."),
			"host":              inputs.NewTagInfo("Hostname."),
			"instance":          inputs.NewTagInfo("Instance."),
			"server_go_version": inputs.NewTagInfo("Server go version."),
			"server_id":         inputs.NewTagInfo("Server ID."),
			"server_version":    inputs.NewTagInfo("Server version."),
			"version":           inputs.NewTagInfo("Version."),
		},
	}
}
