// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package etcd

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	dt "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/prom"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

func TestIntegrate(t *testing.T) {
	if !testutils.CheckIntegrationTestingRunning() {
		t.Skip()
	}

	testutils.PurgeRemoteByName(inputName)       // purge at first.
	defer testutils.PurgeRemoteByName(inputName) // purge at last.

	start := time.Now()
	cases, err := buildCases(t)
	if err != nil {
		cr := &testutils.CaseResult{
			Name:          t.Name(),
			Status:        testutils.TestPassed,
			FailedMessage: err.Error(),
			Cost:          time.Since(start),
		}

		_ = testutils.Flush(cr)
		return
	}

	t.Logf("testing %d cases...", len(cases))

	for _, tc := range cases {
		func(tc *caseSpec) {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				caseStart := time.Now()

				t.Logf("testing %s...", tc.name)

				if err := testutils.RetryTestRun(tc.run); err != nil {
					tc.cr.Status = testutils.TestFailed
					tc.cr.FailedMessage = err.Error()

					panic(err)
				} else {
					tc.cr.Status = testutils.TestPassed
				}

				tc.cr.Cost = time.Since(caseStart)

				require.NoError(t, testutils.Flush(tc.cr))

				t.Cleanup(func() {
					// clean remote docker resources
					if tc.resource == nil {
						return
					}

					tc.pool.Purge(tc.resource)
				})
			})
		}(tc)
	}
}

func getConfAccessPoint(host, port string) string {
	return fmt.Sprintf("http://%s/metrics", net.JoinHostPort(host, port))
}

func buildCases(t *testing.T) ([]*caseSpec, error) {
	t.Helper()

	remote := testutils.GetRemote()

	bases := []struct {
		name         string
		conf         string
		exposedPorts []string
		opts         []inputs.PointCheckOption
	}{
		////////////////////////////////////////////////////////////////////////
		// etcd 3.5.7
		////////////////////////////////////////////////////////////////////////
		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/bitnami/etcd:3.5.7",
			conf: `source = "etcd"
metric_types = ["counter", "gauge"]
interval = "10s"
measurement_name = "etcd"`, // set conf URL later.
			exposedPorts: []string{"2379/tcp"},
			opts: []inputs.PointCheckOption{
				inputs.WithOptionalFields(
					"etcd_cluster_version",
					"etcd_debugging_auth_revision",
					"etcd_debugging_lease_granted_total",
					"etcd_debugging_lease_renewed_total",
					"etcd_debugging_lease_revoked_total",
					"etcd_debugging_mvcc_compact_revision",
					"etcd_debugging_mvcc_current_revision",
					"etcd_debugging_mvcc_db_compaction_keys_total",
					"etcd_debugging_mvcc_db_compaction_last",
					"etcd_debugging_mvcc_db_total_size_in_bytes",
					"etcd_debugging_mvcc_delete_total",
					"etcd_debugging_mvcc_events_total",
					"etcd_debugging_mvcc_keys_total",
					"etcd_debugging_mvcc_pending_events_total",
					"etcd_debugging_mvcc_put_total",
					"etcd_debugging_mvcc_put_total",
					"etcd_debugging_mvcc_range_total",
					"etcd_debugging_mvcc_slow_watcher_total",
					"etcd_debugging_mvcc_total_put_size_in_bytes",
					"etcd_debugging_mvcc_txn_total",
					"etcd_debugging_mvcc_watch_stream_total",
					"etcd_debugging_mvcc_watcher_total",
					"etcd_debugging_server_lease_expired_total",
					"etcd_debugging_store_expires_total",
					"etcd_debugging_store_reads_total",
					"etcd_debugging_store_watch_requests_total",
					"etcd_debugging_store_watchers",
					"etcd_debugging_store_writes_total",
					"etcd_disk_defrag_inflight",
					"etcd_disk_wal_write_bytes_total",
					"etcd_grpc_proxy_cache_hits_total",
					"etcd_grpc_proxy_cache_keys_total",
					"etcd_grpc_proxy_cache_misses_total",
					"etcd_grpc_proxy_events_coalescing_total",
					"etcd_grpc_proxy_watchers_coalescing_total",
					"etcd_mvcc_db_open_read_transactions",
					"etcd_mvcc_db_total_size_in_bytes",
					"etcd_mvcc_db_total_size_in_use_in_bytes",
					"etcd_mvcc_delete_total",
					"etcd_mvcc_put_total",
					"etcd_mvcc_range_total",
					"etcd_mvcc_txn_total",
					"etcd_network_client_grpc_received_bytes_total",
					"etcd_network_client_grpc_sent_bytes_total",
					"etcd_server_go_version",
					"etcd_server_has_leader",
					"etcd_server_health_failures",
					"etcd_server_health_success",
					"etcd_server_heartbeat_send_failures_total",
					"etcd_server_id",
					"etcd_server_is_leader",
					"etcd_server_is_learner",
					"etcd_server_leader_changes_seen_total",
					"etcd_server_learner_promote_successes",
					"etcd_server_proposals_applied_total",
					"etcd_server_proposals_committed_total",
					"etcd_server_proposals_failed_total",
					"etcd_server_proposals_pending",
					"etcd_server_quota_backend_bytes",
					"etcd_server_read_indexes_failed_total",
					"etcd_server_slow_apply_total",
					"etcd_server_slow_read_indexes_total",
					"etcd_server_snapshot_apply_in_progress_total",
					"etcd_server_version",
					"go_goroutines",
					"go_info",
					"go_memstats_alloc_bytes_total",
					"go_memstats_alloc_bytes",
					"go_memstats_buck_hash_sys_bytes",
					"go_memstats_frees_total",
					"go_memstats_gc_cpu_fraction",
					"go_memstats_gc_sys_bytes",
					"go_memstats_heap_alloc_bytes",
					"go_memstats_heap_idle_bytes",
					"go_memstats_heap_inuse_bytes",
					"go_memstats_heap_objects",
					"go_memstats_heap_released_bytes",
					"go_memstats_heap_sys_bytes",
					"go_memstats_last_gc_time_seconds",
					"go_memstats_lookups_total",
					"go_memstats_mallocs_total",
					"go_memstats_mcache_inuse_bytes",
					"go_memstats_mcache_sys_bytes",
					"go_memstats_mspan_inuse_bytes",
					"go_memstats_mspan_sys_bytes",
					"go_memstats_next_gc_bytes",
					"go_memstats_other_sys_bytes",
					"go_memstats_stack_inuse_bytes",
					"go_memstats_stack_sys_bytes",
					"go_memstats_sys_bytes",
					"go_threads",
					"grpc_server_handled_total",
					"grpc_server_msg_received_total",
					"grpc_server_msg_sent_total",
					"grpc_server_started_total",
					"os_fd_limit",
					"os_fd_used",
					"process_cpu_seconds_total",
					"process_max_fds",
					"process_open_fds",
					"process_resident_memory_bytes",
					"process_start_time_seconds",
					"process_virtual_memory_bytes",
					"process_virtual_memory_max_bytes",
					"promhttp_metric_handler_requests_in_flight",
					"promhttp_metric_handler_requests_total",
				),
				inputs.WithOptionalTags(
					"action",
					"cluster_version",
					"code",
					"grpc_code",
					"grpc_method",
					"grpc_service",
					"grpc_type",
					"instance",
					"server_go_version",
					"server_id",
					"server_version",
					"version",
				),
			},
		},

		////////////////////////////////////////////////////////////////////////
		// etcd 3.4.24
		////////////////////////////////////////////////////////////////////////
		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/bitnami/etcd:3.4.24",
			conf: `source = "etcd"
metric_types = ["counter", "gauge"]
interval = "10s"
measurement_name = "etcd"`, // set conf URL later.
			exposedPorts: []string{"2379/tcp"},
			opts: []inputs.PointCheckOption{
				inputs.WithOptionalFields(
					"etcd_cluster_version",
					"etcd_debugging_auth_revision",
					"etcd_debugging_lease_granted_total",
					"etcd_debugging_lease_renewed_total",
					"etcd_debugging_lease_revoked_total",
					"etcd_debugging_mvcc_compact_revision",
					"etcd_debugging_mvcc_current_revision",
					"etcd_debugging_mvcc_db_compaction_keys_total",
					"etcd_debugging_mvcc_db_compaction_last",
					"etcd_debugging_mvcc_db_total_size_in_bytes",
					"etcd_debugging_mvcc_delete_total",
					"etcd_debugging_mvcc_events_total",
					"etcd_debugging_mvcc_keys_total",
					"etcd_debugging_mvcc_pending_events_total",
					"etcd_debugging_mvcc_put_total",
					"etcd_debugging_mvcc_put_total",
					"etcd_debugging_mvcc_range_total",
					"etcd_debugging_mvcc_slow_watcher_total",
					"etcd_debugging_mvcc_total_put_size_in_bytes",
					"etcd_debugging_mvcc_txn_total",
					"etcd_debugging_mvcc_watch_stream_total",
					"etcd_debugging_mvcc_watcher_total",
					"etcd_debugging_server_lease_expired_total",
					"etcd_debugging_store_expires_total",
					"etcd_debugging_store_reads_total",
					"etcd_debugging_store_watch_requests_total",
					"etcd_debugging_store_watchers",
					"etcd_debugging_store_writes_total",
					"etcd_disk_defrag_inflight",
					"etcd_disk_wal_write_bytes_total",
					"etcd_grpc_proxy_cache_hits_total",
					"etcd_grpc_proxy_cache_keys_total",
					"etcd_grpc_proxy_cache_misses_total",
					"etcd_grpc_proxy_events_coalescing_total",
					"etcd_grpc_proxy_watchers_coalescing_total",
					"etcd_mvcc_db_open_read_transactions",
					"etcd_mvcc_db_total_size_in_bytes",
					"etcd_mvcc_db_total_size_in_use_in_bytes",
					"etcd_mvcc_delete_total",
					"etcd_mvcc_put_total",
					"etcd_mvcc_range_total",
					"etcd_mvcc_txn_total",
					"etcd_network_client_grpc_received_bytes_total",
					"etcd_network_client_grpc_sent_bytes_total",
					"etcd_server_go_version",
					"etcd_server_has_leader",
					"etcd_server_health_failures",
					"etcd_server_health_success",
					"etcd_server_heartbeat_send_failures_total",
					"etcd_server_id",
					"etcd_server_is_leader",
					"etcd_server_is_learner",
					"etcd_server_leader_changes_seen_total",
					"etcd_server_learner_promote_successes",
					"etcd_server_proposals_applied_total",
					"etcd_server_proposals_committed_total",
					"etcd_server_proposals_failed_total",
					"etcd_server_proposals_pending",
					"etcd_server_quota_backend_bytes",
					"etcd_server_read_indexes_failed_total",
					"etcd_server_slow_apply_total",
					"etcd_server_slow_read_indexes_total",
					"etcd_server_snapshot_apply_in_progress_total",
					"etcd_server_version",
					"go_goroutines",
					"go_info",
					"go_memstats_alloc_bytes_total",
					"go_memstats_alloc_bytes",
					"go_memstats_buck_hash_sys_bytes",
					"go_memstats_frees_total",
					"go_memstats_gc_cpu_fraction",
					"go_memstats_gc_sys_bytes",
					"go_memstats_heap_alloc_bytes",
					"go_memstats_heap_idle_bytes",
					"go_memstats_heap_inuse_bytes",
					"go_memstats_heap_objects",
					"go_memstats_heap_released_bytes",
					"go_memstats_heap_sys_bytes",
					"go_memstats_last_gc_time_seconds",
					"go_memstats_lookups_total",
					"go_memstats_mallocs_total",
					"go_memstats_mcache_inuse_bytes",
					"go_memstats_mcache_sys_bytes",
					"go_memstats_mspan_inuse_bytes",
					"go_memstats_mspan_sys_bytes",
					"go_memstats_next_gc_bytes",
					"go_memstats_other_sys_bytes",
					"go_memstats_stack_inuse_bytes",
					"go_memstats_stack_sys_bytes",
					"go_memstats_sys_bytes",
					"go_threads",
					"grpc_server_handled_total",
					"grpc_server_msg_received_total",
					"grpc_server_msg_sent_total",
					"grpc_server_started_total",
					"os_fd_limit",
					"os_fd_used",
					"process_cpu_seconds_total",
					"process_max_fds",
					"process_open_fds",
					"process_resident_memory_bytes",
					"process_start_time_seconds",
					"process_virtual_memory_bytes",
					"process_virtual_memory_max_bytes",
					"promhttp_metric_handler_requests_in_flight",
					"promhttp_metric_handler_requests_total",
				),
				inputs.WithOptionalTags(
					"action",
					"cluster_version",
					"code",
					"grpc_code",
					"grpc_method",
					"grpc_service",
					"grpc_type",
					"instance",
					"server_go_version",
					"server_id",
					"server_version",
					"version",
				),
			},
		},

		////////////////////////////////////////////////////////////////////////
		// etcd 3.3.27
		////////////////////////////////////////////////////////////////////////
		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/bitnami/etcd:3.3.27",
			conf: `source = "etcd"
metric_types = ["counter", "gauge"]
interval = "10s"
measurement_name = "etcd"`, // set conf URL later.
			exposedPorts: []string{"2379/tcp"},
			opts: []inputs.PointCheckOption{
				inputs.WithOptionalFields(
					"etcd_cluster_version",
					"etcd_debugging_auth_revision",
					"etcd_debugging_lease_granted_total",
					"etcd_debugging_lease_renewed_total",
					"etcd_debugging_lease_revoked_total",
					"etcd_debugging_mvcc_compact_revision",
					"etcd_debugging_mvcc_current_revision",
					"etcd_debugging_mvcc_db_compaction_keys_total",
					"etcd_debugging_mvcc_db_compaction_last",
					"etcd_debugging_mvcc_db_total_size_in_bytes",
					"etcd_debugging_mvcc_delete_total",
					"etcd_debugging_mvcc_events_total",
					"etcd_debugging_mvcc_keys_total",
					"etcd_debugging_mvcc_pending_events_total",
					"etcd_debugging_mvcc_put_total",
					"etcd_debugging_mvcc_put_total",
					"etcd_debugging_mvcc_range_total",
					"etcd_debugging_mvcc_slow_watcher_total",
					"etcd_debugging_mvcc_total_put_size_in_bytes",
					"etcd_debugging_mvcc_txn_total",
					"etcd_debugging_mvcc_watch_stream_total",
					"etcd_debugging_mvcc_watcher_total",
					"etcd_debugging_server_lease_expired_total",
					"etcd_debugging_store_expires_total",
					"etcd_debugging_store_reads_total",
					"etcd_debugging_store_watch_requests_total",
					"etcd_debugging_store_watchers",
					"etcd_debugging_store_writes_total",
					"etcd_disk_defrag_inflight",
					"etcd_disk_wal_write_bytes_total",
					"etcd_grpc_proxy_cache_hits_total",
					"etcd_grpc_proxy_cache_keys_total",
					"etcd_grpc_proxy_cache_misses_total",
					"etcd_grpc_proxy_events_coalescing_total",
					"etcd_grpc_proxy_watchers_coalescing_total",
					"etcd_mvcc_db_open_read_transactions",
					"etcd_mvcc_db_total_size_in_bytes",
					"etcd_mvcc_db_total_size_in_use_in_bytes",
					"etcd_mvcc_delete_total",
					"etcd_mvcc_put_total",
					"etcd_mvcc_range_total",
					"etcd_mvcc_txn_total",
					"etcd_network_client_grpc_received_bytes_total",
					"etcd_network_client_grpc_sent_bytes_total",
					"etcd_server_go_version",
					"etcd_server_has_leader",
					"etcd_server_health_failures",
					"etcd_server_health_success",
					"etcd_server_heartbeat_send_failures_total",
					"etcd_server_id",
					"etcd_server_is_leader",
					"etcd_server_is_learner",
					"etcd_server_leader_changes_seen_total",
					"etcd_server_learner_promote_successes",
					"etcd_server_proposals_applied_total",
					"etcd_server_proposals_committed_total",
					"etcd_server_proposals_failed_total",
					"etcd_server_proposals_pending",
					"etcd_server_quota_backend_bytes",
					"etcd_server_read_indexes_failed_total",
					"etcd_server_slow_apply_total",
					"etcd_server_slow_read_indexes_total",
					"etcd_server_snapshot_apply_in_progress_total",
					"etcd_server_version",
					"go_goroutines",
					"go_info",
					"go_memstats_alloc_bytes_total",
					"go_memstats_alloc_bytes",
					"go_memstats_buck_hash_sys_bytes",
					"go_memstats_frees_total",
					"go_memstats_gc_cpu_fraction",
					"go_memstats_gc_sys_bytes",
					"go_memstats_heap_alloc_bytes",
					"go_memstats_heap_idle_bytes",
					"go_memstats_heap_inuse_bytes",
					"go_memstats_heap_objects",
					"go_memstats_heap_released_bytes",
					"go_memstats_heap_sys_bytes",
					"go_memstats_last_gc_time_seconds",
					"go_memstats_lookups_total",
					"go_memstats_mallocs_total",
					"go_memstats_mcache_inuse_bytes",
					"go_memstats_mcache_sys_bytes",
					"go_memstats_mspan_inuse_bytes",
					"go_memstats_mspan_sys_bytes",
					"go_memstats_next_gc_bytes",
					"go_memstats_other_sys_bytes",
					"go_memstats_stack_inuse_bytes",
					"go_memstats_stack_sys_bytes",
					"go_memstats_sys_bytes",
					"go_threads",
					"grpc_server_handled_total",
					"grpc_server_msg_received_total",
					"grpc_server_msg_sent_total",
					"grpc_server_started_total",
					"os_fd_limit",
					"os_fd_used",
					"process_cpu_seconds_total",
					"process_max_fds",
					"process_open_fds",
					"process_resident_memory_bytes",
					"process_start_time_seconds",
					"process_virtual_memory_bytes",
					"process_virtual_memory_max_bytes",
					"promhttp_metric_handler_requests_in_flight",
					"promhttp_metric_handler_requests_total",
				),
				inputs.WithOptionalTags(
					"action",
					"cluster_version",
					"code",
					"grpc_code",
					"grpc_method",
					"grpc_service",
					"grpc_type",
					"instance",
					"server_go_version",
					"server_id",
					"server_version",
					"version",
				),
			},
		},
	}

	var cases []*caseSpec

	// compose cases
	for _, base := range bases {
		feeder := io.NewMockedFeeder()

		ipt := prom.NewProm() // This is real prom
		ipt.Feeder = feeder   // Flush metric data to testing_metrics

		// URL from ENV.
		_, err := toml.Decode(base.conf, ipt)
		require.NoError(t, err)

		envs := []string{
			"ALLOW_NONE_AUTHENTICATION=yes",
		}

		repoTag := strings.Split(base.name, ":")

		cases = append(cases, &caseSpec{
			t:       t,
			ipt:     ipt,
			name:    base.name,
			feeder:  feeder,
			envs:    envs,
			repo:    repoTag[0],
			repoTag: repoTag[1],
			opts:    base.opts,

			exposedPorts: base.exposedPorts,

			// Test case result.
			cr: &testutils.CaseResult{
				Name:        t.Name(),
				Case:        base.name,
				ExtraFields: map[string]any{},
				ExtraTags: map[string]string{
					"image":       repoTag[0],
					"image_tag":   repoTag[1],
					"docker_host": remote.Host,
					"docker_port": remote.Port,
				},
			},
		})
	}
	return cases, nil
}

type caseSpec struct {
	t *testing.T

	name         string
	repo         string // docker name
	repoTag      string // docker tag
	envs         []string
	exposedPorts []string
	serverPorts  []string
	opts         []inputs.PointCheckOption
	mCount       map[string]struct{}

	ipt    *prom.Input // This is real prom
	feeder *io.MockedFeeder

	pool     *dt.Pool
	resource *dt.Resource

	cr *testutils.CaseResult // collect `go test -run` metric
}

func (cs *caseSpec) checkPoint(pts []*point.Point) error {
	for _, pt := range pts {
		// fmt.Printf("pt = %s\n", pt.LineProto())

		var opts []inputs.PointCheckOption
		opts = append(opts, inputs.WithExtraTags(cs.ipt.Tags))

		measurement := string(pt.Name())

		switch measurement {
		case inputName:
			opts = append(opts, cs.opts...)
			opts = append(opts, inputs.WithDoc(&etcdMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[inputName] = struct{}{}

		default: // TODO: check other measurement
			panic("unknown measurement: " + measurement)
		}

		// check if tag appended
		if len(cs.ipt.Tags) != 0 {
			// cs.t.Logf("checking tags %+#v...", cs.ipt.Tags)

			tags := pt.Tags()
			for k, expect := range cs.ipt.Tags {
				if v := tags.Get([]byte(k)); v != nil {
					got := string(v.GetD())
					if got != expect {
						return fmt.Errorf("expect tag value %s, got %s", expect, got)
					}
				} else {
					return fmt.Errorf("tag %s not found, got %v", k, tags)
				}
			}
		}
	}

	// TODO: some other checking on @pts, such as `if some required measurements exist'...

	return nil
}

func (cs *caseSpec) run() error {
	// start remote image server
	r := testutils.GetRemote()
	dockerTCP := r.TCPURL() // got "tcp://" + net.JoinHostPort(i.Host, i.Port) 2375

	cs.t.Logf("get remote: %+#v, TCP: %s", r, dockerTCP)

	start := time.Now()

	p, err := dt.NewPool(dockerTCP)
	if err != nil {
		return err
	}

	uniqueContainerName := testutils.GetUniqueContainerName(inputName)

	resource, err := p.RunWithOptions(&dt.RunOptions{
		Name: uniqueContainerName, // ATTENTION: not cs.name.

		// specify container image & tag
		Repository: cs.repo,
		Tag:        cs.repoTag,

		ExposedPorts: cs.exposedPorts,

		// container run-time envs
		Env: cs.envs,
	}, func(c *docker.HostConfig) {
		c.RestartPolicy = docker.RestartPolicy{Name: "no"}
		c.AutoRemove = true
	})
	if err != nil {
		return err
	}

	cs.pool = p
	cs.resource = resource

	if err := cs.getMappingPorts(); err != nil {
		return err
	}
	cs.ipt.URL = getConfAccessPoint(r.Host, cs.serverPorts[0]) // set conf URL here.

	cs.t.Logf("check service(%s:%v)...", r.Host, cs.serverPorts)

	if err := cs.portsOK(r); err != nil {
		return err
	}

	cs.cr.AddField("container_ready_cost", int64(time.Since(start)))

	var wg sync.WaitGroup

	// start input
	cs.t.Logf("start input...")
	wg.Add(1)
	go func() {
		defer wg.Done()
		cs.ipt.Run()
	}()

	// wait data
	start = time.Now()
	cs.t.Logf("wait points...")
	pts, err := cs.feeder.AnyPoints()
	if err != nil {
		return err
	}

	cs.cr.AddField("point_latency", int64(time.Since(start)))
	cs.cr.AddField("point_count", len(pts))

	cs.t.Logf("get %d points", len(pts))
	cs.mCount = make(map[string]struct{})
	if err := cs.checkPoint(pts); err != nil {
		return err
	}

	cs.t.Logf("stop input...")
	cs.ipt.Terminate()

	require.Equal(cs.t, 1, len(cs.mCount)) // Metric set count.

	cs.t.Logf("exit...")
	wg.Wait()

	return nil
}

func (cs *caseSpec) getMappingPorts() error {
	cs.serverPorts = make([]string, len(cs.exposedPorts))
	for k, v := range cs.exposedPorts {
		mapStr := cs.resource.GetHostPort(v)
		_, port, err := net.SplitHostPort(mapStr)
		if err != nil {
			return err
		}
		cs.serverPorts[k] = port
	}
	return nil
}

func (cs *caseSpec) portsOK(r *testutils.RemoteInfo) error {
	for _, v := range cs.serverPorts {
		if !r.PortOK(docker.Port(v).Port(), time.Minute) {
			return fmt.Errorf("service checking failed")
		}
	}
	return nil
}
