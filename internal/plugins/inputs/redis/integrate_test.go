// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	dockertest "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

// ATTENTION: Docker version should use v20.10.18 in integrate tests. Other versions are not tested.

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

func buildCases(t *testing.T) ([]*caseSpec, error) {
	t.Helper()

	remote := testutils.GetRemote()

	bases := []struct {
		name                 string // Also used as build image name:tag.
		conf                 string
		dockerFileText       string // Empty if not build image.
		exposedPorts         []string
		cmd                  []string
		optsRedisBigkey      []inputs.PointCheckOption
		optsRedisClient      []inputs.PointCheckOption
		optsRedisCluster     []inputs.PointCheckOption
		optsRedisCommandStat []inputs.PointCheckOption
		optsRedisDB          []inputs.PointCheckOption
		optsRedisInfoM       []inputs.PointCheckOption
		optsRedisReplica     []inputs.PointCheckOption
	}{
		////////////////////////////////////////////////////////////////////////
		// redis:4.0.14
		////////////////////////////////////////////////////////////////////////
		{
			// with tags
			name: "redis:4.0.14-alpine",
			conf: `interval = "2s"
		           slow_log = true
		           all_slow_log = false
		           slowlog-max-len = 128
		           election = true
			       [tags]
			         foo = "bar"`, // set conf URL later.
			exposedPorts: []string{"6379/tcp"},
			optsRedisBigkey: []inputs.PointCheckOption{
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1", "foo": "bar"}),
			},
			optsRedisClient: []inputs.PointCheckOption{
				inputs.WithOptionalFields("id", "fd", "age", "idle", "db", "sub", "psub", "ssub", "multi", "qbuf", "qbuf_free", "argv_mem", "multi_mem", "obl", "oll", "omem", "tot_mem", "redir", "resp"),
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1", "foo": "bar"}),
			},
			optsRedisCluster: []inputs.PointCheckOption{
				inputs.WithOptionalFields("cluster_state", "cluster_slots_assigned", "cluster_slots_ok", "cluster_slots_pfail", "cluster_slots_fail", "cluster_known_nodes", "cluster_size", "cluster_current_epoch", "cluster_my_epoch", "cluster_stats_messages_sent", "cluster_stats_messages_received", "total_cluster_links_buffer_limit_exceeded", "cluster_stats_messages_ping_sent", "cluster_stats_messages_ping_received", "cluster_stats_messages_pong_sent", "cluster_stats_messages_pong_received", "cluster_stats_messages_meet_sent", "cluster_stats_messages_meet_received", "cluster_stats_messages_fail_sent", "cluster_stats_messages_fail_received", "cluster_stats_messages_publish_sent", "cluster_stats_messages_publish_received", "cluster_stats_messages_auth_req_sent", "cluster_stats_messages_auth_req_received", "cluster_stats_messages_auth_ack_sent", "cluster_stats_messages_auth_ack_received", "cluster_stats_messages_update_sent", "cluster_stats_messages_update_received", "cluster_stats_messages_mfstart_sent", "cluster_stats_messages_mfstart_received", "cluster_stats_messages_module_sent", "cluster_stats_messages_module_received", "cluster_stats_messages_publishshard_sent", "cluster_stats_messages_publishshard_received"),
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1", "foo": "bar"}),
			},
			optsRedisCommandStat: []inputs.PointCheckOption{
				inputs.WithOptionalFields("calls", "usec", "usec_per_call", "rejected_calls", "failed_calls"),
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1", "foo": "bar"}),
			},
			optsRedisDB: []inputs.PointCheckOption{
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1", "foo": "bar"}),
			},
			optsRedisInfoM: []inputs.PointCheckOption{
				inputs.WithOptionalFields(getInfoField()...),
				inputs.WithOptionalTags("server", "service_name", "command_type", "error_type", "quantile"),
				inputs.WithExtraTags(map[string]string{"election": "1", "foo": "bar"}),
			},
			optsRedisReplica: []inputs.PointCheckOption{
				inputs.WithOptionalFields("master_link_down_since_seconds"),
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1", "foo": "bar"}),
			},
		},
		{
			name: "redis:4.0.14-alpine",
			conf: `interval = "2s"
					slow_log = true
					all_slow_log = false
					slowlog-max-len = 128
					election = true`, // set conf URL later.
			exposedPorts: []string{"6379/tcp"},
			optsRedisBigkey: []inputs.PointCheckOption{
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
			optsRedisClient: []inputs.PointCheckOption{
				inputs.WithOptionalFields("id", "fd", "age", "idle", "db", "sub", "psub", "ssub", "multi", "qbuf", "qbuf_free", "argv_mem", "multi_mem", "obl", "oll", "omem", "tot_mem", "redir", "resp"),
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
			optsRedisCluster: []inputs.PointCheckOption{
				inputs.WithOptionalFields("cluster_state", "cluster_slots_assigned", "cluster_slots_ok", "cluster_slots_pfail", "cluster_slots_fail", "cluster_known_nodes", "cluster_size", "cluster_current_epoch", "cluster_my_epoch", "cluster_stats_messages_sent", "cluster_stats_messages_received", "total_cluster_links_buffer_limit_exceeded", "cluster_stats_messages_ping_sent", "cluster_stats_messages_ping_received", "cluster_stats_messages_pong_sent", "cluster_stats_messages_pong_received", "cluster_stats_messages_meet_sent", "cluster_stats_messages_meet_received", "cluster_stats_messages_fail_sent", "cluster_stats_messages_fail_received", "cluster_stats_messages_publish_sent", "cluster_stats_messages_publish_received", "cluster_stats_messages_auth_req_sent", "cluster_stats_messages_auth_req_received", "cluster_stats_messages_auth_ack_sent", "cluster_stats_messages_auth_ack_received", "cluster_stats_messages_update_sent", "cluster_stats_messages_update_received", "cluster_stats_messages_mfstart_sent", "cluster_stats_messages_mfstart_received", "cluster_stats_messages_module_sent", "cluster_stats_messages_module_received", "cluster_stats_messages_publishshard_sent", "cluster_stats_messages_publishshard_received"),
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
			optsRedisCommandStat: []inputs.PointCheckOption{
				inputs.WithOptionalFields("calls", "usec", "usec_per_call", "rejected_calls", "failed_calls"),
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
			optsRedisDB: []inputs.PointCheckOption{
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
			optsRedisInfoM: []inputs.PointCheckOption{
				inputs.WithOptionalFields(getInfoField()...),
				inputs.WithOptionalTags("server", "service_name", "command_type", "error_type", "quantile"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
			optsRedisReplica: []inputs.PointCheckOption{
				inputs.WithOptionalFields("master_link_down_since_seconds"),
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
		},
		{
			name: "redis:4.0.14-alpine",
			conf: `interval = "2s"
					slow_log = true
					all_slow_log = false
					slowlog-max-len = 128
					election = false`, // set conf URL later.
			exposedPorts: []string{"6379/tcp"},
			optsRedisBigkey: []inputs.PointCheckOption{
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisClient: []inputs.PointCheckOption{
				inputs.WithOptionalFields("id", "fd", "age", "idle", "db", "sub", "psub", "ssub", "multi", "qbuf", "qbuf_free", "argv_mem", "multi_mem", "obl", "oll", "omem", "tot_mem", "redir", "resp"),
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisCluster: []inputs.PointCheckOption{
				inputs.WithOptionalFields("cluster_state", "cluster_slots_assigned", "cluster_slots_ok", "cluster_slots_pfail", "cluster_slots_fail", "cluster_known_nodes", "cluster_size", "cluster_current_epoch", "cluster_my_epoch", "cluster_stats_messages_sent", "cluster_stats_messages_received", "total_cluster_links_buffer_limit_exceeded", "cluster_stats_messages_ping_sent", "cluster_stats_messages_ping_received", "cluster_stats_messages_pong_sent", "cluster_stats_messages_pong_received", "cluster_stats_messages_meet_sent", "cluster_stats_messages_meet_received", "cluster_stats_messages_fail_sent", "cluster_stats_messages_fail_received", "cluster_stats_messages_publish_sent", "cluster_stats_messages_publish_received", "cluster_stats_messages_auth_req_sent", "cluster_stats_messages_auth_req_received", "cluster_stats_messages_auth_ack_sent", "cluster_stats_messages_auth_ack_received", "cluster_stats_messages_update_sent", "cluster_stats_messages_update_received", "cluster_stats_messages_mfstart_sent", "cluster_stats_messages_mfstart_received", "cluster_stats_messages_module_sent", "cluster_stats_messages_module_received", "cluster_stats_messages_publishshard_sent", "cluster_stats_messages_publishshard_received"),
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisCommandStat: []inputs.PointCheckOption{
				inputs.WithOptionalFields("calls", "usec", "usec_per_call", "rejected_calls", "failed_calls"),
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisDB: []inputs.PointCheckOption{
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisInfoM: []inputs.PointCheckOption{
				inputs.WithOptionalFields(getInfoField()...),
				inputs.WithOptionalTags("server", "service_name", "command_type", "error_type", "quantile"),
			},
			optsRedisReplica: []inputs.PointCheckOption{
				inputs.WithOptionalFields("master_link_down_since_seconds"),
				inputs.WithOptionalTags("service_name"),
			},
		},

		////////////////////////////////////////////////////////////////////////
		// redis:5.0.14
		////////////////////////////////////////////////////////////////////////
		{
			name: "redis:5.0.14-alpine",
			conf: `interval = "2s"
					slow_log = true
					all_slow_log = false
					slowlog-max-len = 128
					election = true`, // set conf URL later.
			exposedPorts: []string{"6379/tcp"},
			optsRedisBigkey: []inputs.PointCheckOption{
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
			optsRedisClient: []inputs.PointCheckOption{
				inputs.WithOptionalFields("id", "fd", "age", "idle", "db", "sub", "psub", "ssub", "multi", "qbuf", "qbuf_free", "argv_mem", "multi_mem", "obl", "oll", "omem", "tot_mem", "redir", "resp"),
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
			optsRedisCluster: []inputs.PointCheckOption{
				inputs.WithOptionalFields("cluster_state", "cluster_slots_assigned", "cluster_slots_ok", "cluster_slots_pfail", "cluster_slots_fail", "cluster_known_nodes", "cluster_size", "cluster_current_epoch", "cluster_my_epoch", "cluster_stats_messages_sent", "cluster_stats_messages_received", "total_cluster_links_buffer_limit_exceeded", "cluster_stats_messages_ping_sent", "cluster_stats_messages_ping_received", "cluster_stats_messages_pong_sent", "cluster_stats_messages_pong_received", "cluster_stats_messages_meet_sent", "cluster_stats_messages_meet_received", "cluster_stats_messages_fail_sent", "cluster_stats_messages_fail_received", "cluster_stats_messages_publish_sent", "cluster_stats_messages_publish_received", "cluster_stats_messages_auth_req_sent", "cluster_stats_messages_auth_req_received", "cluster_stats_messages_auth_ack_sent", "cluster_stats_messages_auth_ack_received", "cluster_stats_messages_update_sent", "cluster_stats_messages_update_received", "cluster_stats_messages_mfstart_sent", "cluster_stats_messages_mfstart_received", "cluster_stats_messages_module_sent", "cluster_stats_messages_module_received", "cluster_stats_messages_publishshard_sent", "cluster_stats_messages_publishshard_received"),
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
			optsRedisCommandStat: []inputs.PointCheckOption{
				inputs.WithOptionalFields("calls", "usec", "usec_per_call", "rejected_calls", "failed_calls"),
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
			optsRedisDB: []inputs.PointCheckOption{
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
			optsRedisInfoM: []inputs.PointCheckOption{
				inputs.WithOptionalFields(getInfoField()...),
				inputs.WithOptionalTags("server", "service_name", "command_type", "error_type", "quantile"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
			optsRedisReplica: []inputs.PointCheckOption{
				inputs.WithOptionalFields("master_link_down_since_seconds"),
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
		},
		{
			name: "redis:5.0.14-alpine",
			conf: `interval = "2s"
					slow_log = true
					all_slow_log = false
					slowlog-max-len = 128
					election = false`, // set conf URL later.
			exposedPorts: []string{"6379/tcp"},
			optsRedisBigkey: []inputs.PointCheckOption{
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisClient: []inputs.PointCheckOption{
				inputs.WithOptionalFields("id", "fd", "age", "idle", "db", "sub", "psub", "ssub", "multi", "qbuf", "qbuf_free", "argv_mem", "multi_mem", "obl", "oll", "omem", "tot_mem", "redir", "resp"),
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisCluster: []inputs.PointCheckOption{
				inputs.WithOptionalFields("cluster_state", "cluster_slots_assigned", "cluster_slots_ok", "cluster_slots_pfail", "cluster_slots_fail", "cluster_known_nodes", "cluster_size", "cluster_current_epoch", "cluster_my_epoch", "cluster_stats_messages_sent", "cluster_stats_messages_received", "total_cluster_links_buffer_limit_exceeded", "cluster_stats_messages_ping_sent", "cluster_stats_messages_ping_received", "cluster_stats_messages_pong_sent", "cluster_stats_messages_pong_received", "cluster_stats_messages_meet_sent", "cluster_stats_messages_meet_received", "cluster_stats_messages_fail_sent", "cluster_stats_messages_fail_received", "cluster_stats_messages_publish_sent", "cluster_stats_messages_publish_received", "cluster_stats_messages_auth_req_sent", "cluster_stats_messages_auth_req_received", "cluster_stats_messages_auth_ack_sent", "cluster_stats_messages_auth_ack_received", "cluster_stats_messages_update_sent", "cluster_stats_messages_update_received", "cluster_stats_messages_mfstart_sent", "cluster_stats_messages_mfstart_received", "cluster_stats_messages_module_sent", "cluster_stats_messages_module_received", "cluster_stats_messages_publishshard_sent", "cluster_stats_messages_publishshard_received"),
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisCommandStat: []inputs.PointCheckOption{
				inputs.WithOptionalFields("calls", "usec", "usec_per_call", "rejected_calls", "failed_calls"),
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisDB: []inputs.PointCheckOption{
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisInfoM: []inputs.PointCheckOption{
				inputs.WithOptionalFields(getInfoField()...),
				inputs.WithOptionalTags("server", "service_name", "command_type", "error_type", "quantile"),
			},
			optsRedisReplica: []inputs.PointCheckOption{
				inputs.WithOptionalFields("master_link_down_since_seconds"),
				inputs.WithOptionalTags("service_name"),
			},
		},

		////////////////////////////////////////////////////////////////////////
		// redis:6.2.12
		////////////////////////////////////////////////////////////////////////
		{
			name: "redis:6.2.12-alpine",
			conf: `interval = "2s"
					slow_log = true
					all_slow_log = false
					slowlog-max-len = 128
					election = true`, // set conf URL later.
			exposedPorts: []string{"6379/tcp"},
			optsRedisBigkey: []inputs.PointCheckOption{
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
			optsRedisClient: []inputs.PointCheckOption{
				inputs.WithOptionalFields("id", "fd", "age", "idle", "db", "sub", "psub", "ssub", "multi", "qbuf", "qbuf_free", "argv_mem", "multi_mem", "obl", "oll", "omem", "tot_mem", "redir", "resp"),
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
			optsRedisCluster: []inputs.PointCheckOption{
				inputs.WithOptionalFields("cluster_state", "cluster_slots_assigned", "cluster_slots_ok", "cluster_slots_pfail", "cluster_slots_fail", "cluster_known_nodes", "cluster_size", "cluster_current_epoch", "cluster_my_epoch", "cluster_stats_messages_sent", "cluster_stats_messages_received", "total_cluster_links_buffer_limit_exceeded", "cluster_stats_messages_ping_sent", "cluster_stats_messages_ping_received", "cluster_stats_messages_pong_sent", "cluster_stats_messages_pong_received", "cluster_stats_messages_meet_sent", "cluster_stats_messages_meet_received", "cluster_stats_messages_fail_sent", "cluster_stats_messages_fail_received", "cluster_stats_messages_publish_sent", "cluster_stats_messages_publish_received", "cluster_stats_messages_auth_req_sent", "cluster_stats_messages_auth_req_received", "cluster_stats_messages_auth_ack_sent", "cluster_stats_messages_auth_ack_received", "cluster_stats_messages_update_sent", "cluster_stats_messages_update_received", "cluster_stats_messages_mfstart_sent", "cluster_stats_messages_mfstart_received", "cluster_stats_messages_module_sent", "cluster_stats_messages_module_received", "cluster_stats_messages_publishshard_sent", "cluster_stats_messages_publishshard_received"),
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
			optsRedisCommandStat: []inputs.PointCheckOption{
				inputs.WithOptionalFields("calls", "usec", "usec_per_call", "rejected_calls", "failed_calls"),
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
			optsRedisDB: []inputs.PointCheckOption{
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
			optsRedisInfoM: []inputs.PointCheckOption{
				inputs.WithOptionalFields(getInfoField()...),
				inputs.WithOptionalTags("server", "service_name", "command_type", "error_type", "quantile"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
			optsRedisReplica: []inputs.PointCheckOption{
				inputs.WithOptionalFields("master_link_down_since_seconds"),
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
		},
		{
			name: "redis:6.2.12-alpine",
			conf: `interval = "2s"
					slow_log = true
					all_slow_log = false
					slowlog-max-len = 128
					election = false`, // set conf URL later.
			exposedPorts: []string{"6379/tcp"},
			optsRedisBigkey: []inputs.PointCheckOption{
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisClient: []inputs.PointCheckOption{
				inputs.WithOptionalFields("id", "fd", "age", "idle", "db", "sub", "psub", "ssub", "multi", "qbuf", "qbuf_free", "argv_mem", "multi_mem", "obl", "oll", "omem", "tot_mem", "redir", "resp"),
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisCluster: []inputs.PointCheckOption{
				inputs.WithOptionalFields("cluster_state", "cluster_slots_assigned", "cluster_slots_ok", "cluster_slots_pfail", "cluster_slots_fail", "cluster_known_nodes", "cluster_size", "cluster_current_epoch", "cluster_my_epoch", "cluster_stats_messages_sent", "cluster_stats_messages_received", "total_cluster_links_buffer_limit_exceeded", "cluster_stats_messages_ping_sent", "cluster_stats_messages_ping_received", "cluster_stats_messages_pong_sent", "cluster_stats_messages_pong_received", "cluster_stats_messages_meet_sent", "cluster_stats_messages_meet_received", "cluster_stats_messages_fail_sent", "cluster_stats_messages_fail_received", "cluster_stats_messages_publish_sent", "cluster_stats_messages_publish_received", "cluster_stats_messages_auth_req_sent", "cluster_stats_messages_auth_req_received", "cluster_stats_messages_auth_ack_sent", "cluster_stats_messages_auth_ack_received", "cluster_stats_messages_update_sent", "cluster_stats_messages_update_received", "cluster_stats_messages_mfstart_sent", "cluster_stats_messages_mfstart_received", "cluster_stats_messages_module_sent", "cluster_stats_messages_module_received", "cluster_stats_messages_publishshard_sent", "cluster_stats_messages_publishshard_received"),
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisCommandStat: []inputs.PointCheckOption{
				inputs.WithOptionalFields("calls", "usec", "usec_per_call", "rejected_calls", "failed_calls"),
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisDB: []inputs.PointCheckOption{
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisInfoM: []inputs.PointCheckOption{
				inputs.WithOptionalFields(getInfoField()...),
				inputs.WithOptionalTags("server", "service_name", "command_type", "error_type", "quantile"),
			},
			optsRedisReplica: []inputs.PointCheckOption{
				inputs.WithOptionalFields("master_link_down_since_seconds"),
				inputs.WithOptionalTags("service_name"),
			},
		},

		////////////////////////////////////////////////////////////////////////
		// redis:7.0.11
		////////////////////////////////////////////////////////////////////////
		{
			name: "redis:7.0.11-alpine",
			conf: `interval = "2s"
					slow_log = true
					all_slow_log = false
					slowlog-max-len = 128
					election = true`, // set conf URL later.
			exposedPorts: []string{"6379/tcp"},
			optsRedisBigkey: []inputs.PointCheckOption{
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
			optsRedisClient: []inputs.PointCheckOption{
				inputs.WithOptionalFields("id", "fd", "age", "idle", "db", "sub", "psub", "ssub", "multi", "qbuf", "qbuf_free", "argv_mem", "multi_mem", "obl", "oll", "omem", "tot_mem", "redir", "resp"),
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
			optsRedisCluster: []inputs.PointCheckOption{
				inputs.WithOptionalFields("cluster_state", "cluster_slots_assigned", "cluster_slots_ok", "cluster_slots_pfail", "cluster_slots_fail", "cluster_known_nodes", "cluster_size", "cluster_current_epoch", "cluster_my_epoch", "cluster_stats_messages_sent", "cluster_stats_messages_received", "total_cluster_links_buffer_limit_exceeded", "cluster_stats_messages_ping_sent", "cluster_stats_messages_ping_received", "cluster_stats_messages_pong_sent", "cluster_stats_messages_pong_received", "cluster_stats_messages_meet_sent", "cluster_stats_messages_meet_received", "cluster_stats_messages_fail_sent", "cluster_stats_messages_fail_received", "cluster_stats_messages_publish_sent", "cluster_stats_messages_publish_received", "cluster_stats_messages_auth_req_sent", "cluster_stats_messages_auth_req_received", "cluster_stats_messages_auth_ack_sent", "cluster_stats_messages_auth_ack_received", "cluster_stats_messages_update_sent", "cluster_stats_messages_update_received", "cluster_stats_messages_mfstart_sent", "cluster_stats_messages_mfstart_received", "cluster_stats_messages_module_sent", "cluster_stats_messages_module_received", "cluster_stats_messages_publishshard_sent", "cluster_stats_messages_publishshard_received"),
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
			optsRedisCommandStat: []inputs.PointCheckOption{
				inputs.WithOptionalFields("calls", "usec", "usec_per_call", "rejected_calls", "failed_calls"),
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
			optsRedisDB: []inputs.PointCheckOption{
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
			optsRedisInfoM: []inputs.PointCheckOption{
				inputs.WithOptionalFields(getInfoField()...),
				inputs.WithOptionalTags("server", "service_name", "command_type", "error_type", "quantile"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
			optsRedisReplica: []inputs.PointCheckOption{
				inputs.WithOptionalFields("master_link_down_since_seconds"),
				inputs.WithOptionalTags("service_name"),
				inputs.WithExtraTags(map[string]string{"election": "1"}),
			},
		},
		{
			name: "redis:7.0.11-alpine",
			conf: `interval = "2s"
				slow_log = true
				all_slow_log = false
				slowlog-max-len = 128
				election = false`, // set conf URL later.
			exposedPorts: []string{"6379/tcp"},
			optsRedisBigkey: []inputs.PointCheckOption{
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisClient: []inputs.PointCheckOption{
				inputs.WithOptionalFields("id", "fd", "age", "idle", "db", "sub", "psub", "ssub", "multi", "qbuf", "qbuf_free", "argv_mem", "multi_mem", "obl", "oll", "omem", "tot_mem", "redir", "resp"),
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisCluster: []inputs.PointCheckOption{
				inputs.WithOptionalFields("cluster_state", "cluster_slots_assigned", "cluster_slots_ok", "cluster_slots_pfail", "cluster_slots_fail", "cluster_known_nodes", "cluster_size", "cluster_current_epoch", "cluster_my_epoch", "cluster_stats_messages_sent", "cluster_stats_messages_received", "total_cluster_links_buffer_limit_exceeded", "cluster_stats_messages_ping_sent", "cluster_stats_messages_ping_received", "cluster_stats_messages_pong_sent", "cluster_stats_messages_pong_received", "cluster_stats_messages_meet_sent", "cluster_stats_messages_meet_received", "cluster_stats_messages_fail_sent", "cluster_stats_messages_fail_received", "cluster_stats_messages_publish_sent", "cluster_stats_messages_publish_received", "cluster_stats_messages_auth_req_sent", "cluster_stats_messages_auth_req_received", "cluster_stats_messages_auth_ack_sent", "cluster_stats_messages_auth_ack_received", "cluster_stats_messages_update_sent", "cluster_stats_messages_update_received", "cluster_stats_messages_mfstart_sent", "cluster_stats_messages_mfstart_received", "cluster_stats_messages_module_sent", "cluster_stats_messages_module_received", "cluster_stats_messages_publishshard_sent", "cluster_stats_messages_publishshard_received"),
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisCommandStat: []inputs.PointCheckOption{
				inputs.WithOptionalFields("calls", "usec", "usec_per_call", "rejected_calls", "failed_calls"),
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisDB: []inputs.PointCheckOption{
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisInfoM: []inputs.PointCheckOption{
				inputs.WithOptionalFields(getInfoField()...),
				inputs.WithOptionalTags("server", "service_name", "command_type", "error_type", "quantile"),
			},
			optsRedisReplica: []inputs.PointCheckOption{
				inputs.WithOptionalFields("master_link_down_since_seconds"),
				inputs.WithOptionalTags("service_name"),
			},
		},
		{
			name: "redis:7.0.11-alpine",
			conf: `interval = "2s"
				slow_log = true
				all_slow_log = false
				slowlog-max-len = 128
				election = false
				latency_percentiles = true`, // set conf URL later.
			exposedPorts: []string{"6379/tcp"},
			optsRedisBigkey: []inputs.PointCheckOption{
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisClient: []inputs.PointCheckOption{
				inputs.WithOptionalFields("id", "fd", "age", "idle", "db", "sub", "psub", "ssub", "multi", "qbuf", "qbuf_free", "argv_mem", "multi_mem", "obl", "oll", "omem", "tot_mem", "redir", "resp"),
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisCluster: []inputs.PointCheckOption{
				inputs.WithOptionalFields("cluster_state", "cluster_slots_assigned", "cluster_slots_ok", "cluster_slots_pfail", "cluster_slots_fail", "cluster_known_nodes", "cluster_size", "cluster_current_epoch", "cluster_my_epoch", "cluster_stats_messages_sent", "cluster_stats_messages_received", "total_cluster_links_buffer_limit_exceeded", "cluster_stats_messages_ping_sent", "cluster_stats_messages_ping_received", "cluster_stats_messages_pong_sent", "cluster_stats_messages_pong_received", "cluster_stats_messages_meet_sent", "cluster_stats_messages_meet_received", "cluster_stats_messages_fail_sent", "cluster_stats_messages_fail_received", "cluster_stats_messages_publish_sent", "cluster_stats_messages_publish_received", "cluster_stats_messages_auth_req_sent", "cluster_stats_messages_auth_req_received", "cluster_stats_messages_auth_ack_sent", "cluster_stats_messages_auth_ack_received", "cluster_stats_messages_update_sent", "cluster_stats_messages_update_received", "cluster_stats_messages_mfstart_sent", "cluster_stats_messages_mfstart_received", "cluster_stats_messages_module_sent", "cluster_stats_messages_module_received", "cluster_stats_messages_publishshard_sent", "cluster_stats_messages_publishshard_received"),
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisCommandStat: []inputs.PointCheckOption{
				inputs.WithOptionalFields("calls", "usec", "usec_per_call", "rejected_calls", "failed_calls"),
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisDB: []inputs.PointCheckOption{
				inputs.WithOptionalTags("service_name"),
			},
			optsRedisInfoM: []inputs.PointCheckOption{
				inputs.WithOptionalFields(getInfoField()...),
				inputs.WithOptionalTags("server", "service_name", "command_type", "error_type", "quantile"),
			},
			optsRedisReplica: []inputs.PointCheckOption{
				inputs.WithOptionalFields("master_link_down_since_seconds"),
				inputs.WithOptionalTags("service_name"),
			},
		},
	}

	var cases []*caseSpec

	// compose cases
	for _, base := range bases {
		feeder := io.NewMockedFeeder()

		ipt := defaultInput()
		ipt.feeder = feeder

		_, err := toml.Decode(base.conf, ipt)
		require.NoError(t, err)

		if ipt.Election {
			ipt.tagger = testutils.NewTaggerElection()
		} else {
			ipt.tagger = testutils.NewTaggerHost()
		}

		repoTag := strings.Split(base.name, ":")

		cases = append(cases, &caseSpec{
			t:       t,
			ipt:     ipt,
			name:    base.name,
			feeder:  feeder,
			repo:    repoTag[0],
			repoTag: repoTag[1],

			dockerFileText: base.dockerFileText,
			exposedPorts:   base.exposedPorts,
			cmd:            base.cmd,

			optsRedisBigkey:      base.optsRedisBigkey,
			optsRedisClient:      base.optsRedisClient,
			optsRedisCluster:     base.optsRedisCluster,
			optsRedisCommandStat: base.optsRedisCommandStat,
			optsRedisDB:          base.optsRedisDB,
			optsRedisInfoM:       base.optsRedisInfoM,
			optsRedisReplica:     base.optsRedisReplica,

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

////////////////////////////////////////////////////////////////////////////////

// caseSpec.

type caseSpec struct {
	t *testing.T

	name                 string
	repo                 string
	repoTag              string
	dockerFileText       string
	exposedPorts         []string
	serverPorts          []string
	optsRedisBigkey      []inputs.PointCheckOption
	optsRedisClient      []inputs.PointCheckOption
	optsRedisCluster     []inputs.PointCheckOption
	optsRedisCommandStat []inputs.PointCheckOption
	optsRedisDB          []inputs.PointCheckOption
	optsRedisInfoM       []inputs.PointCheckOption
	optsRedisReplica     []inputs.PointCheckOption
	cmd                  []string
	mCount               map[string]struct{}

	ipt    *Input
	feeder *io.MockedFeeder

	pool     *dockertest.Pool
	resource *dockertest.Resource

	cr *testutils.CaseResult
}

func (cs *caseSpec) checkPoint(pts []*point.Point) error {
	for _, pt := range pts {
		var opts []inputs.PointCheckOption

		measurement := pt.Name()

		switch measurement {
		case redisBigkey:
			opts = append(opts, cs.optsRedisBigkey...)
			opts = append(opts, inputs.WithDoc(&bigKeyMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[redisBigkey] = struct{}{}

		case redisClient:
			opts = append(opts, cs.optsRedisClient...)
			opts = append(opts, inputs.WithDoc(&clientMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[redisClient] = struct{}{}

		case redisCluster:
			opts = append(opts, cs.optsRedisCluster...)
			opts = append(opts, inputs.WithDoc(&clusterMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[redisCluster] = struct{}{}

		case redisCommandStat:
			opts = append(opts, cs.optsRedisCommandStat...)
			opts = append(opts, inputs.WithDoc(&commandMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[redisCommandStat] = struct{}{}

		case redisDB:
			opts = append(opts, cs.optsRedisDB...)
			opts = append(opts, inputs.WithDoc(&dbMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[redisDB] = struct{}{}

		case redisInfoM:
			opts = append(opts, cs.optsRedisInfoM...)
			opts = append(opts, inputs.WithDoc(&infoMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[redisInfoM] = struct{}{}

		case redisReplica:
			opts = append(opts, cs.optsRedisReplica...)
			opts = append(opts, inputs.WithDoc(&replicaMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[redisReplica] = struct{}{}

		default: // TODO: check other measurement
			panic("unknown measurement: " + measurement)
		}

		// check if tag appended
		if len(cs.ipt.Tags) != 0 {
			cs.t.Logf("%s checking tags %+#v...", measurement, cs.ipt.Tags)

			tags := pt.Tags()
			for k, expect := range cs.ipt.Tags {
				if v := tags.Get(k); v != nil {
					got := v.GetS()
					if got != expect {
						return fmt.Errorf("%s expect tag value %s, got %s", measurement, expect, got)
					}
				} else {
					return fmt.Errorf("%s tag %s not found, got %v", measurement, k, tags)
				}
			}
		}
	}

	// TODO: some other checking on @pts, such as `if some required measurements exist'...

	return nil
}

func (cs *caseSpec) run() error {
	r := testutils.GetRemote()
	dockerTCP := r.TCPURL()

	cs.t.Logf("get remote: %+#v, TCP: %s", r, dockerTCP)

	start := time.Now()

	p, err := cs.getPool(dockerTCP)
	if err != nil {
		return err
	}

	dockerFileDir, dockerFilePath, err := cs.getDockerFilePath()
	if err != nil {
		return err
	}
	defer os.RemoveAll(dockerFileDir)

	uniqueContainerName := testutils.GetUniqueContainerName(inputName)

	var resource *dockertest.Resource

	if len(cs.dockerFileText) == 0 {
		// Just run a container from existing docker image.
		resource, err = p.RunWithOptions(
			&dockertest.RunOptions{
				Name: uniqueContainerName, // ATTENTION: not cs.name.

				Repository: cs.repo,
				Tag:        cs.repoTag,
				Cmd:        cs.cmd,

				ExposedPorts: cs.exposedPorts,
			},

			func(c *docker.HostConfig) {
				c.RestartPolicy = docker.RestartPolicy{Name: "no"}
				c.AutoRemove = true
			},
		)
	} else {
		// Build docker image from Dockerfile and run a container from it.
		resource, err = p.BuildAndRunWithOptions(
			dockerFilePath,

			&dockertest.RunOptions{
				ContainerName: uniqueContainerName,
				Name:          cs.name, // ATTENTION: not uniqueContainerName.

				Repository: cs.repo,
				Tag:        cs.repoTag,
				Cmd:        cs.cmd,

				ExposedPorts: cs.exposedPorts,
			},

			func(c *docker.HostConfig) {
				c.RestartPolicy = docker.RestartPolicy{Name: "no"}
				c.AutoRemove = true
			},
		)
	}

	if err != nil {
		return err
	}

	cs.pool = p
	cs.resource = resource

	if err := cs.getMappingPorts(); err != nil {
		return err
	}
	cs.ipt.Host = r.Host
	cs.ipt.Port, err = strconv.Atoi(cs.serverPorts[0])
	if err != nil {
		return err
	}

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
	// pts, err := cs.feeder.NPoints(60, 5*time.Minute)
	pts, err := cs.feeder.NPoints(20, 5*time.Minute)
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

	require.GreaterOrEqual(cs.t, len(cs.mCount), 2) // At lest 2 Metric out.

	cs.t.Logf("exit...")
	wg.Wait()

	return nil
}

func (cs *caseSpec) getPool(endpoint string) (*dockertest.Pool, error) {
	p, err := dockertest.NewPool(endpoint)
	if err != nil {
		return nil, err
	}
	err = p.Client.Ping()
	if err != nil {
		cs.t.Logf("Could not connect to Docker: %v", err)
		return nil, err
	}
	return p, nil
}

func (cs *caseSpec) getDockerFilePath() (dirName string, fileName string, err error) {
	if len(cs.dockerFileText) == 0 {
		return
	}

	tmpDir, err := ioutil.TempDir("", "dockerfiles_")
	if err != nil {
		cs.t.Logf("ioutil.TempDir failed: %s", err.Error())
		return "", "", err
	}

	tmpFile, err := ioutil.TempFile(tmpDir, "dockerfile_")
	if err != nil {
		cs.t.Logf("ioutil.TempFile failed: %s", err.Error())
		return "", "", err
	}

	_, err = tmpFile.WriteString(cs.dockerFileText)
	if err != nil {
		cs.t.Logf("TempFile.WriteString failed: %s", err.Error())
		return "", "", err
	}

	if err := os.Chmod(tmpFile.Name(), os.ModePerm); err != nil {
		cs.t.Logf("os.Chmod failed: %s", err.Error())
		return "", "", err
	}

	if err := tmpFile.Close(); err != nil {
		cs.t.Logf("Close failed: %s", err.Error())
		return "", "", err
	}

	return tmpDir, tmpFile.Name(), nil
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

////////////////////////////////////////////////////////////////////////////////
