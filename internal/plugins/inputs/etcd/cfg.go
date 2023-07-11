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
			"etcd_debugging_mvcc_db_compaction_keys_total":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging mvcc db compaction keys total."},
			"etcd_debugging_mvcc_db_compaction_last":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging mvcc db compaction last."},
			"etcd_debugging_mvcc_db_total_size_in_bytes":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging mvcc db total size in bytes."},
			"etcd_debugging_mvcc_delete_total":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging mvcc delete total."},
			"etcd_debugging_mvcc_events_total":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging mvcc events total."},
			"etcd_debugging_mvcc_keys_total":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging mvcc keys total."},
			"etcd_debugging_mvcc_pending_events_total":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging mvcc pending events total."},
			"etcd_debugging_mvcc_put_total":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging mvcc put total."},
			"etcd_debugging_mvcc_range_total":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging mvcc range total."},
			"etcd_debugging_mvcc_slow_watcher_total":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging mvcc slow watcher total."},
			"etcd_debugging_mvcc_total_put_size_in_bytes":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging mvcc total put size in bytes."},
			"etcd_debugging_mvcc_txn_total":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging mvcc txn total."},
			"etcd_debugging_mvcc_watch_stream_total":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging mvcc watch stream total."},
			"etcd_debugging_mvcc_watcher_total":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging mvcc watcher total."},
			"etcd_debugging_server_lease_expired_total":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging server lease expired total"},
			"etcd_debugging_store_expires_total":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging store expires total."},
			"etcd_debugging_store_reads_total":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging store reads total."},
			"etcd_debugging_store_watch_requests_total":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging store watch requests total"},
			"etcd_debugging_store_watchers":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging store watchers"},
			"etcd_debugging_store_writes_total":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd debugging store writes total"},
			"etcd_disk_defrag_inflight":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd disk defrag inflight"},
			"etcd_disk_wal_write_bytes_total":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd disk wal write bytes total."},
			"etcd_grpc_proxy_cache_hits_total":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd grpc proxy cache hits total."},
			"etcd_grpc_proxy_cache_keys_total":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd grpc proxy cache keys total."},
			"etcd_grpc_proxy_cache_misses_total":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd grpc proxy cache misses total."},
			"etcd_grpc_proxy_events_coalescing_total":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd grpc proxy events coalescing total."},
			"etcd_grpc_proxy_watchers_coalescing_total":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd grpc proxy watchers coalescing total."},
			"etcd_mvcc_db_open_read_transactions":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd mvcc db open read transactions."},
			"etcd_mvcc_db_total_size_in_bytes":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd mvcc db total size in bytes."},
			"etcd_mvcc_db_total_size_in_use_in_bytes":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd mvcc db total size in use in bytes."},
			"etcd_mvcc_delete_total":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd mvcc delete total."},
			"etcd_mvcc_put_total":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd mvcc put total."},
			"etcd_mvcc_range_total":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd mvcc range total."},
			"etcd_mvcc_txn_total":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd mvcc txn total."},
			"etcd_network_client_grpc_received_bytes_total": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd network client grpc received bytes total."},
			"etcd_network_client_grpc_sent_bytes_total":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd network client grpc sent bytes total."},
			"etcd_server_go_version":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server go version."},
			"etcd_server_has_leader":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server has leader."},
			"etcd_server_health_failures":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server health failures."},
			"etcd_server_health_success":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server health success."},
			"etcd_server_heartbeat_send_failures_total":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server heartbeat send failures total."},
			"etcd_server_id":                                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server id."},
			"etcd_server_is_leader":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server is leader."},
			"etcd_server_is_learner":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server is learner."},
			"etcd_server_leader_changes_seen_total":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server leader changes seen total."},
			"etcd_server_learner_promote_successes":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server learner promote successes."},
			"etcd_server_proposals_applied_total":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server proposals applied total."},
			"etcd_server_proposals_committed_total":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server proposals committed total."},
			"etcd_server_proposals_failed_total":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server proposals failed total."},
			"etcd_server_proposals_pending":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server proposals pending."},
			"etcd_server_quota_backend_bytes":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server quota backend bytes."},
			"etcd_server_read_indexes_failed_total":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server read indexes failed total."},
			"etcd_server_slow_apply_total":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server slow apply total."},
			"etcd_server_slow_read_indexes_total":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server slow read indexes total."},
			"etcd_server_snapshot_apply_in_progress_total":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server snapshot apply in progress total."},
			"etcd_server_version":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "etcd server version."},
			"go_goroutines":                                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go goroutines."},
			"go_info":                                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go info."},
			"go_memstats_alloc_bytes_total":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go memstats alloc bytes total."},
			"go_memstats_alloc_bytes":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go memstats alloc bytes."},
			"go_memstats_buck_hash_sys_bytes":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go memstats buck hash sys bytes."},
			"go_memstats_frees_total":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go memstats frees total."},
			"go_memstats_gc_cpu_fraction":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go memstats gc cpu fraction."},
			"go_memstats_gc_sys_bytes":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go memstats gc sys bytes."},
			"go_memstats_heap_alloc_bytes":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go memstats heap alloc bytes."},
			"go_memstats_heap_idle_bytes":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go memstats heap idle bytes."},
			"go_memstats_heap_inuse_bytes":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go memstats heap inuse bytes."},
			"go_memstats_heap_objects":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go memstats heap objects."},
			"go_memstats_heap_released_bytes":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go memstats heap released bytes."},
			"go_memstats_heap_sys_bytes":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go memstats heap sys bytes."},
			"go_memstats_last_gc_time_seconds":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go memstats last gc time seconds."},
			"go_memstats_lookups_total":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go memstats lookups total."},
			"go_memstats_mallocs_total":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go memstats mallocs total."},
			"go_memstats_mcache_inuse_bytes":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go memstats mcache inuse bytes."},
			"go_memstats_mcache_sys_bytes":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go memstats mcache sys bytes."},
			"go_memstats_mspan_inuse_bytes":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go memstats mspan inuse bytes."},
			"go_memstats_mspan_sys_bytes":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go memstats mspan sys bytes."},
			"go_memstats_next_gc_bytes":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go memstats next gc bytes."},
			"go_memstats_other_sys_bytes":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go memstats other sys bytes."},
			"go_memstats_stack_inuse_bytes":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go memstats stack inuse bytes."},
			"go_memstats_stack_sys_bytes":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go memstats stack sys bytes."},
			"go_memstats_sys_bytes":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go memstats sys bytes."},
			"go_threads":                                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Go threads."},
			"grpc_server_handled_total":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "GRPC server handled total."},
			"grpc_server_msg_received_total":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "GRPC server msg received total."},
			"grpc_server_msg_sent_total":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "GRPC server msg sent total."},
			"grpc_server_started_total":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "GRPC server started total."},
			"os_fd_limit":                                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "OS file descriptor limit."},
			"os_fd_used":                                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "OS file descriptor used."},
			"process_cpu_seconds_total":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Process cpu seconds total."},
			"process_max_fds":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Process max fds."},
			"process_open_fds":                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Process open fds."},
			"process_resident_memory_bytes":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Process resident memory bytes."},
			"process_start_time_seconds":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Process start time in seconds."},
			"process_virtual_memory_bytes":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Process virtual memory bytes."},
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
