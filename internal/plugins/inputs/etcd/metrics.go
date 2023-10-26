// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package etcd

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var _ inputs.Measurement = (*etcdMeasurement)(nil)

type etcdMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// Point implement MeasurementV2.
func (m *etcdMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *etcdMeasurement) Info() *inputs.MeasurementInfo {
	fields := internal.CopyMapStringInterface(etcdFields)

	inputs.AppendGolangGeneralFields(fields)

	return &inputs.MeasurementInfo{
		Name:   inputName,
		Desc:   "etcd metrics.",
		Fields: fields,
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

////////////////////////////////////////////////////////////////////////////////

//nolint:lll
var etcdFields = map[string]interface{}{
	"etcd_cluster_version":                                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Which version is running. 1 for 'cluster_version' label with current cluster version"},
	"etcd_debugging_auth_revision":                                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The current revision of auth store."},
	"etcd_debugging_disk_backend_commit_rebalance_duration_seconds":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Histogram, Unit: inputs.NCount, Desc: "The latency distributions of commit.rebalance called by bboltdb backend."},
	"etcd_debugging_disk_backend_commit_spill_duration_seconds":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Histogram, Unit: inputs.NCount, Desc: "The latency distributions of commit.spill called by bboltdb backend."},
	"etcd_debugging_disk_backend_commit_write_duration_seconds":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Histogram, Unit: inputs.NCount, Desc: "The latency distributions of commit.write called by bboltdb backend."},
	"etcd_debugging_lease_granted_total":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of granted leases."},
	"etcd_debugging_lease_renewed_total":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of renewed leases seen by the leader."},
	"etcd_debugging_lease_revoked_total":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of revoked leases."},
	"etcd_debugging_lease_ttl_total":                                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Histogram, Unit: inputs.NCount, Desc: "Bucketed histogram of lease TTLs."},
	"etcd_debugging_mvcc_compact_revision":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The revision of the last compaction in store."},
	"etcd_debugging_mvcc_current_revision":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The current revision of store."},
	"etcd_debugging_mvcc_db_compaction_keys_total":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of db keys compacted."},
	"etcd_debugging_mvcc_db_compaction_last":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The unix time of the last db compaction. Resets to 0 on start."},
	"etcd_debugging_mvcc_db_compaction_pause_duration_milliseconds":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Histogram, Unit: inputs.NCount, Desc: "Bucketed histogram of db compaction pause duration."},
	"etcd_debugging_mvcc_db_compaction_total_duration_milliseconds":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Histogram, Unit: inputs.NCount, Desc: "Bucketed histogram of db compaction total duration."},
	"etcd_debugging_mvcc_db_total_size_in_bytes":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total size of the underlying database physically allocated in bytes."},
	"etcd_debugging_mvcc_delete_total":                                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of deletes seen by this member."},
	"etcd_debugging_mvcc_events_total":                                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of events sent by this member."},
	"etcd_debugging_mvcc_index_compaction_pause_duration_milliseconds": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Histogram, Unit: inputs.NCount, Desc: "Bucketed histogram of index compaction pause duration."},
	"etcd_debugging_mvcc_keys_total":                                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of keys."},
	"etcd_debugging_mvcc_pending_events_total":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of pending events to be sent."},
	"etcd_debugging_mvcc_put_total":                                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of puts seen by this member."},
	"etcd_debugging_mvcc_range_total":                                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of ranges seen by this member."},
	"etcd_debugging_mvcc_slow_watcher_total":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of unsynced slow watchers."},
	"etcd_debugging_mvcc_total_put_size_in_bytes":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total size of put kv pairs seen by this member."},
	"etcd_debugging_mvcc_txn_total":                                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of txns seen by this member."},
	"etcd_debugging_mvcc_watch_stream_total":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of watch streams."},
	"etcd_debugging_mvcc_watcher_total":                                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of watchers."},
	"etcd_debugging_server_alarms":                                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Alarms for every member in cluster. 1 for 'server_id' label with current ID. 2 for 'alarm_type' label with type of this alarm"},
	"etcd_debugging_server_lease_expired_total":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of expired leases."},
	"etcd_debugging_snap_save_total_duration_seconds":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Histogram, Unit: inputs.NCount, Desc: "The total latency distributions of save called by snapshot."},
	"etcd_debugging_snap_save_marshalling_duration_seconds":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Histogram, Unit: inputs.NCount, Desc: "The marshaling cost distributions of save called by snapshot."}, //nolint:misspell
	"etcd_debugging_store_expires_total":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of expired keys."},
	"etcd_debugging_store_reads_failed_total":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Failed read actions by (get/getRecursive), local to this member."},
	"etcd_debugging_store_reads_total":                                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of reads action by (get/getRecursive), local to this member."},
	"etcd_debugging_store_watch_requests_total":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of incoming watch requests (new or reestablished)."},
	"etcd_debugging_store_watchers":                                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Count of currently active watchers."},
	"etcd_debugging_store_writes_failed_total":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Failed write actions (e.g. set/compareAndDelete), seen by this member."},
	"etcd_debugging_store_writes_total":                                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of writes (e.g. set/compareAndDelete) seen by this member."},
	"etcd_disk_backend_commit_duration_seconds":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Histogram, Unit: inputs.NCount, Desc: "The latency distributions of commit called by backend."},
	"etcd_disk_backend_defrag_duration_seconds":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Histogram, Unit: inputs.NCount, Desc: "The latency distribution of backend defragmentation."},
	"etcd_disk_backend_snapshot_duration_seconds":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Histogram, Unit: inputs.NCount, Desc: "The latency distribution of backend snapshots."},
	"etcd_disk_defrag_inflight":                                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Whether or not defrag is active on the member. 1 means active, 0 means not."},
	"etcd_disk_wal_fsync_duration_seconds":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Histogram, Unit: inputs.NCount, Desc: "The latency distributions of fsync called by WAL."},
	"etcd_disk_wal_write_bytes_total":                                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of bytes written in WAL."},
	"etcd_grpc_proxy_cache_hits_total":                                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of cache hits"},
	"etcd_grpc_proxy_cache_keys_total":                                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of keys/ranges cached"},
	"etcd_grpc_proxy_cache_misses_total":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of cache misses"},
	"etcd_grpc_proxy_events_coalescing_total":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of events coalescing"},
	"etcd_grpc_proxy_watchers_coalescing_total":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of current watchers coalescing"},
	"etcd_mvcc_db_open_read_transactions":                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of currently open read transactions"},
	"etcd_mvcc_db_total_size_in_bytes":                                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total size of the underlying database physically allocated in bytes."},
	"etcd_mvcc_db_total_size_in_use_in_bytes":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total size of the underlying database logically in use in bytes."},
	"etcd_mvcc_delete_total":                                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of deletes seen by this member."},
	"etcd_mvcc_hash_duration_seconds":                                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Histogram, Unit: inputs.NCount, Desc: "The latency distribution of storage hash operation."},
	"etcd_mvcc_hash_rev_duration_seconds":                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Histogram, Unit: inputs.NCount, Desc: "The latency distribution of storage hash by revision operation."},
	"etcd_mvcc_put_total":                                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of puts seen by this member."},
	"etcd_mvcc_range_total":                                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of ranges seen by this member."},
	"etcd_mvcc_txn_total":                                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of txns seen by this member."},
	"etcd_network_active_peers":                                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The current number of active peer connections."},
	"etcd_network_client_grpc_received_bytes_total":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of bytes received from grpc clients."},
	"etcd_network_client_grpc_sent_bytes_total":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of bytes sent to grpc clients."},
	"etcd_network_disconnected_peers_total":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of disconnected peers."},
	"etcd_network_known_peers":                                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The current number of known peers."},
	"etcd_network_peer_received_bytes_total":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of bytes received from peers."},
	"etcd_network_peer_received_failures_total":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of receive failures from peers."},
	"etcd_network_peer_round_trip_time_seconds":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Histogram, Unit: inputs.NCount, Desc: "Round-Trip-Time histogram between peers"},
	"etcd_network_peer_sent_bytes_total":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of bytes sent to peers."},
	"etcd_network_peer_sent_failures_total":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of send failures from peers."},
	"etcd_network_server_stream_failures_total":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of stream failures from the local server."},
	"etcd_network_snapshot_receive_failures":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of snapshot receive failures"},
	"etcd_network_snapshot_receive_inflights_total":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of inflight snapshot receives"},
	"etcd_network_snapshot_receive_success":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of successful snapshot receives"},
	"etcd_network_snapshot_receive_total_duration_seconds":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Histogram, Unit: inputs.NCount, Desc: "Total latency distributions of v3 snapshot receives"},
	"etcd_network_snapshot_send_failures":                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of snapshot send failures"},
	"etcd_network_snapshot_send_inflights_total":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of inflight snapshot sends"},
	"etcd_network_snapshot_send_success":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of successful snapshot sends"},
	"etcd_network_snapshot_send_total_duration_seconds":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Histogram, Unit: inputs.NCount, Desc: "Total latency distributions of v3 snapshot sends"},
	"etcd_server_apply_duration_seconds":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Histogram, Unit: inputs.NCount, Desc: "The latency distributions of v2 apply called by backend."},
	"etcd_server_client_requests_total":                                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of client requests per client version."},
	"etcd_server_go_version":                                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Which Go version server is running with. 1 for 'server_go_version' label with current version."},
	"etcd_server_has_leader":                                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Whether or not a leader exists. 1 is existence, 0 is not."},
	"etcd_server_health_failures":                                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of failed health checks"},
	"etcd_server_health_success":                                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of successful health checks"},
	"etcd_server_heartbeat_send_failures_total":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of leader heartbeat send failures (likely overloaded from slow disk)."},
	"etcd_server_id":                                                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Server or member ID in hexadecimal format. 1 for 'server_id' label with current ID."},
	"etcd_server_is_leader":                                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Whether or not this member is a leader. 1 if is, 0 otherwise."},
	"etcd_server_is_learner":                                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Whether or not this member is a learner. 1 if is, 0 otherwise."},
	"etcd_server_leader_changes_seen_total":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of leader changes seen."},
	"etcd_server_learner_promote_failures":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of failed learner promotions (likely learner not ready) while this member is leader."},
	"etcd_server_learner_promote_successes":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of successful learner promotions while this member is leader."},
	"etcd_server_proposals_applied_total":                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of consensus proposals applied."},
	"etcd_server_proposals_committed_total":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of consensus proposals committed."},
	"etcd_server_proposals_failed_total":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of failed proposals seen."},
	"etcd_server_proposals_pending":                                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The current number of pending proposals to commit."},
	"etcd_server_quota_backend_bytes":                                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current backend storage quota size in bytes."},
	"etcd_server_read_indexes_failed_total":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of failed read indexes seen."},
	"etcd_server_slow_apply_total":                                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of slow apply requests (likely overloaded from slow disk)."},
	"etcd_server_slow_read_indexes_total":                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of pending read indexes not in sync with leader's or timed out read index requests."},
	"etcd_server_snapshot_apply_in_progress_total":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "1 if the server is applying the incoming snapshot. 0 if none."},
	"etcd_server_version":                                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Which version is running. 1 for 'server_version' label with current version."},
	"etcd_snap_db_fsync_duration_seconds":                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Histogram, Unit: inputs.NCount, Desc: "The latency distributions of fsyncing .snap.db file"},
	"etcd_snap_db_save_total_duration_seconds":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Histogram, Unit: inputs.NCount, Desc: "The total latency distributions of v3 snapshot save"},
	"etcd_snap_fsync_duration_seconds":                                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Histogram, Unit: inputs.NCount, Desc: "The latency distributions of fsync called by snap."},
	"os_fd_limit":                                                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The file descriptor limit."},
	"os_fd_used":                                                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of used file descriptors."},
}
