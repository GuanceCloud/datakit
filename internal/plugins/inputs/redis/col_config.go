// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	CREDENTIALSTR = "******"
)

func (i *instance) collectConfig(ctx context.Context) {
	collectStart := time.Now()
	allConf, err := i.curCli.configGet(ctx, "*")
	if err != nil {
		l.Error("redis exec command `config get *`: %s", err)
		return
	}

	l.Debugf("config get *:\n%+#v", allConf)

	pts := i.parseConfigAll(allConf)

	if err := i.ipt.feeder.Feed(point.Logging, pts,
		dkio.WithCollectCost(time.Since(collectStart)),
		dkio.WithElection(i.ipt.Election),
		dkio.WithSource(dkio.FeedSource(inputName, "config")),
	); err != nil {
		l.Warnf("feed measurement: %s, ignored", err)
	}
}

func (i *instance) parseConfigAll(conf map[string]string) []*point.Point {
	var (
		pts []*point.Point
		kvs point.KVs
	)

	for rawkey, strVal := range conf {
		if strVal == "" {
			strVal = "not-set"
		}

		n, err := strconv.Atoi(strVal)

		// map yes/no to 1/-1
		switch strVal {
		case "yes":
			n = 1
		case "no":
			n = -1
		}

		key := strings.ReplaceAll(rawkey, "-", "_")

		switch rawkey {
		case "client-output-buffer-limit":
			pts = append(pts, i.parseOBLConf(strVal)...)
		case "requirepass",
			"masterauth",
			"tls-key-file-pass",
			"tls-client-key-file-pass":
			// Mask credential fields
			if strVal != "not-set" {
				kvs = kvs.Add(key, CREDENTIALSTR)
			} else {
				kvs = kvs.Add(key, strVal)
			}
		default:
			// All config items as fields
			if err != nil && n == 0 { // parse failed and n not set to 1/-1
				kvs = kvs.Add(key, strVal)
			} else {
				kvs = kvs.Set(key, n)
			}
		}
	}

	opts := append(point.DefaultLoggingOptions(), point.WithTime(i.ipt.ptsTime))
	pts = append(pts, point.NewPoint(measureuemtRedisConfig, kvs, opts...))

	for _, pt := range pts {
		for k, v := range i.mergedTags {
			pt.AddTag(k, v)
		}
	}

	return pts
}

func (i *instance) parseOBLConf(s string) []*point.Point {
	var (
		arr  = strings.Split(s, " ")
		pts  []*point.Point
		opts = append(point.DefaultLoggingOptions(), point.WithTime(i.ipt.ptsTime))
	)

	// example: normal 0 0 0 slave 268435456 67108864 60 pubsub 33554432 8388608 60
	for i := 0; i < len(arr); i += 4 {
		class := arr[i]
		if v, err := strconv.ParseFloat(arr[i+1], 64); err == nil {
			var kvs point.KVs
			kvs = kvs.Add("client_output_buffer_limit_bytes", v).AddTag("class", class+".hard")
			pts = append(pts, point.NewPoint(measureuemtRedisConfig, kvs, opts...))
		}

		if v, err := strconv.ParseFloat(arr[i+2], 64); err == nil {
			var kvs point.KVs
			kvs = kvs.Add("client_output_buffer_limit_bytes", v).AddTag("class", class+".soft")
			pts = append(pts, point.NewPoint(measureuemtRedisConfig, kvs, opts...))
		}

		if v, err := strconv.ParseFloat(arr[i+3], 64); err == nil {
			var kvs point.KVs
			kvs = kvs.Add("client_output_buffer_limit_overcome_seconds", v).AddTag("class", class+".soft")
			pts = append(pts, point.NewPoint(measureuemtRedisConfig, kvs, opts...))
		}
	}

	return pts
}

type configMeasurement struct{}

//nolint:funlen,lll
func (configMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   measureuemtRedisConfig,
		Desc:   "Redis configuration metrics comes from command `CONFIG GET *`(only part of tags and fields are listed)",
		DescZh: "通过执行 `CONFIG GET *` 获取到的 Redis 配置信息（此处仅展示部分指标）",
		Cat:    point.Logging,
		Tags: map[string]interface{}{
			"host":         &inputs.TagInfo{Desc: "Hostname."},
			"server":       &inputs.TagInfo{Desc: "Server addr."},
			"service_name": &inputs.TagInfo{Desc: "Service name."},
			"class": &inputs.TagInfo{
				Desc: "Client output buffer limits on different class(`nomal:hard/normal:soft/slave:hard/slave:soft/pubsub:hard/pubsub:soft`)",
			},
			"appendfilename":      &inputs.TagInfo{Desc: "AOF filename(such as *appendonly.aof*)"},
			"pidfile":             &inputs.TagInfo{Desc: "File path where Redis writes its process ID"},
			"replica_announce_ip": &inputs.TagInfo{Desc: "IP address replica announces to master"},
			"bind":                &inputs.TagInfo{Desc: "Bind Redis to specific network interfaces"},
			"tls_cert_file":       &inputs.TagInfo{Desc: "Path to TLS certificate file"},
			"dbfilename":          &inputs.TagInfo{Desc: "RDB snapshot filename(such as *dump.rdb*)"},
			"aclfile":             &inputs.TagInfo{Desc: "Path to Redis ACL configuration file"},
			"unixsocket":          &inputs.TagInfo{Desc: "Path for Unix domain socket"},
			"tls_key_file":        &inputs.TagInfo{Desc: "Path to TLS private key file"},
			"maxmemory_policy": &inputs.TagInfo{
				Desc: "Key eviction policy when maxmemory is reached(such as `allkeys-lfu`)",
			},
			"dir":                    &inputs.TagInfo{Desc: "Working directory for persistence files"},
			"cluster_announce_ip":    &inputs.TagInfo{Desc: "IP address to announce in cluster"},
			"appendfsync":            &inputs.TagInfo{Desc: "AOF fsync policy (`always/everysec/no`)"},
			"syslog_ident":           &inputs.TagInfo{Desc: "Syslog identity"},
			"tls_ca_cert_file":       &inputs.TagInfo{Desc: "Path to TLS CA certificate bundle"},
			"syslog_facility":        &inputs.TagInfo{Desc: "Syslog facility"},
			"notify_keyspace_events": &inputs.TagInfo{Desc: "Keyspace event notification options"},
			"cluster_config_file":    &inputs.TagInfo{Desc: "Cluster configuration file"},
			"save": &inputs.TagInfo{
				Desc: "RDB snapshot conditions (time/changes), such as `3600 1 300 100 60 10000`",
			},
		},

		Fields: map[string]interface{}{
			"client_output_buffer_limit_bytes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.BytesPerSec,
				Desc:     "Client output buffer limits",
			},

			"client_output_buffer_limit_overcome_seconds": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationSecond,
				Desc:     "Client output buffer soft limit duration",
			},

			"maxmemory_eviction_tenacity": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "Aggressiveness of eviction process (0-100)",
			},

			"appendonly": &inputs.FieldInfo{
				DataType: inputs.Bool,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "Enable Append-Only File persistence",
			},

			"active_defrag_ignore_bytes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "Minimum memory fragmentation to start defrag",
			},

			"min_replicas_max_lag": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationSecond,
				Desc:     "Maximum lag for `min_replicas`",
			},

			"hash_max_ziplist_entries": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Maximum entries for ziplist encoding of hashes",
			},

			"cluster_migration_barrier": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "Minimum replicas for master migration",
			},

			"active_defrag_threshold_upper": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.Percent,
				DataType: inputs.Int,
				Desc:     "Maximum fragmentation percentage to use maximum effort",
			},
			"maxmemory": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				DataType: inputs.Int,
				Desc:     "Maximum memory limit (0=unlimited)",
			},
			"rdb_save_incremental_fsync": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Incremental fsync during RDB saves",
			},
			"tcp_backlog": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				DataType: inputs.Int,
				Desc:     "TCP connection queue size",
			},
			"maxclients": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				DataType: inputs.Int,
				Desc:     "Maximum number of client connections",
			},
			"repl_diskless_sync": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Use diskless replication",
			},
			"dynamic_hz": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Dynamically adjust hz based on clients",
			},
			"lazyfree_lazy_expire": &inputs.FieldInfo{
				Type: inputs.Gauge, Unit: inputs.NoUnit, DataType: inputs.Bool, Desc: "Use lazy freeing for expired keys",
			},
			"auto_aof_rewrite_percentage": &inputs.FieldInfo{
				Type: inputs.Gauge, Unit: inputs.Percent, DataType: inputs.Int, Desc: "AOF rewrite trigger based on size growth",
			},
			"tcp_keepalive": &inputs.FieldInfo{
				Type: inputs.Gauge, Unit: inputs.DurationSecond, DataType: inputs.Int, Desc: "TCP keepalive interval",
			},

			"slowlog_max_len": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				DataType: inputs.Int,
				Desc:     "Maximum number of slow log entries to retain",
			},
			"repl_timeout": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.DurationSecond,
				DataType: inputs.Int,
				Desc:     "Timeout for replication in seconds",
			},
			"replica_ignore_disk_write_errors": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				DataType: inputs.Bool,
				Desc:     "Whether replicas should ignore disk write errors (1=yes, -1=no)",
			},
			"cluster_replica_validity_factor": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Int,
				Desc:     "Factor used to determine if a replica is valid for failover",
			},
			"repl_diskless_sync_max_replicas": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				DataType: inputs.Int,
				Desc:     "Maximum number of replicas to support with diskless sync",
			},
			"latency_monitor_threshold": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				DataType: inputs.Int,
				Desc:     "Threshold in milliseconds for latency monitoring (0=disabled)",
			},
			"cluster_slave_validity_factor": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Int,
				Desc:     "Deprecated alias for `cluster_replica_validity_factor`",
			},
			"aof_use_rdb_preamble": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				DataType: inputs.Bool,
				Desc:     "Whether to use RDB preamble in AOF rewrites (1=yes, -1=no)",
			},
			"slave_serve_stale_data": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				DataType: inputs.Bool,
				Desc:     "Whether replicas should serve stale data when master is down (1=yes, -1=no)",
			},
			"cluster_allow_replica_migration": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				DataType: inputs.Bool,
				Desc:     "Whether cluster allows replica migration between masters (1=yes, -1=no)",
			},
			"zset_max_ziplist_value": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				DataType: inputs.Int,
				Desc:     "Maximum value size in bytes for ziplist encoding of sorted sets",
			},
			"repl_disable_tcp_nodelay": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				DataType: inputs.Bool,
				Desc:     "Disable `TCP_NODELAY` on replica sockets (-1=no, 1=yes)",
			},
			"active_expire_effort": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Int,
				Desc:     "Effort level for active key expiration (1-10)",
			},
			"zset_max_listpack_entries": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				DataType: inputs.Int,
				Desc:     "Maximum entries for listpack encoding of sorted sets",
			},
			"repl_backlog_size": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				DataType: inputs.Int,
				Desc:     "Size in bytes of replication backlog buffer",
			},
			"slowlog_log_slower_than": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				DataType: inputs.Int,
				Desc:     "Threshold in microseconds for slow log entries",
			},
			"set_max_listpack_value": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				DataType: inputs.Int,
				Desc:     "Maximum value size in bytes for listpack encoding of sets",
			},
			"stream_node_max_entries": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				DataType: inputs.Int,
				Desc:     "Maximum number of entries in a single stream node",
			},
			"repl_ping_replica_period": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.DurationSecond,
				DataType: inputs.Int,
				Desc:     "Interval in seconds for pinging replicas",
			},
			"tls_session_cache_size": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				DataType: inputs.Int,
				Desc:     "Size of TLS session cache",
			},
			"aof_timestamp_enabled": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Whether to add timestamps to AOF entries (1=yes, -1=no)",
			},
			"maxmemory_clients": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				DataType: inputs.Int,
				Desc:     "Maximum memory allocated for client buffers",
			},
			"acllog_max_len": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				DataType: inputs.Int,
				Desc:     "Maximum number of ACL log entries to retain",
			},
			"lua_time_limit": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				DataType: inputs.Int,
				Desc:     "Maximum execution time in milliseconds for Lua scripts",
			},
			"hll_sparse_max_bytes": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				DataType: inputs.Int,
				Desc:     "Maximum size in bytes for sparse HyperLogLog representation",
			},
			"io_threads": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				DataType: inputs.Int,
				Desc:     "Number of I/O threads (1=main thread only)",
			},
			"stop_writes_on_bgsave_error": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Stop accepting writes if RDB save fails (1=yes, -1=no)",
			},
			"set_max_intset_entries": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				DataType: inputs.Int,
				Desc:     "Maximum entries for intset encoding of sets",
			},
			"lazyfree_lazy_server_del": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Use lazy freeing for server operations (1=yes, -1=no)",
			},
			"crash_memcheck_enabled": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Enable memory checking on crash (1=yes, -1=no)",
			},
			"repl_backlog_ttl": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.DurationSecond,
				DataType: inputs.Int,
				Desc:     "Time in seconds to retain replication backlog after master disconnect",
			},
			"slave_announce_port": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Int,
				Desc:     "Deprecated alias for `replica_announce_port`",
			},
			"min_slaves_to_write": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				DataType: inputs.Int,
				Desc:     "Deprecated alias for `min_replicas_to_write`",
			},
			"min_replicas_to_write": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				DataType: inputs.Int,
				Desc:     "Minimum number of connected replicas to allow writes",
			},
			"tls_prefer_server_ciphers": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Prefer server's cipher suite order (1=yes, -1=no)",
			},
			"aof_rewrite_incremental_fsync": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Enable incremental fsync during AOF rewrite (1=yes, -1=no)",
			},
			"daemonize": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Run Redis as a daemon (1=yes, -1=no)",
			},
			"jemalloc_bg_thread": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Enable jemalloc background threads (1=yes, -1=no)",
			},
			"replica_read_only": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Replicas accept read-only commands (1=yes, -1=no)",
			},
			"stream_node_max_bytes": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				DataType: inputs.Int,
				Desc:     "Maximum size in bytes for a single stream node",
			},
			"replica_announced": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Whether the replica is announced to the cluster (1=yes, -1=no)",
			},
			"active_defrag_threshold_lower": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.Percent,
				DataType: inputs.Float,
				Desc:     "Minimum memory fragmentation percentage to start active defrag",
			},
			"cluster_announce_tls_port": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Int,
				Desc:     "TLS port that cluster nodes announce to other nodes",
			},
			"tracking_table_max_keys": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				DataType: inputs.Int,
				Desc:     "Maximum number of keys in client tracking table",
			},
			"cluster_require_full_coverage": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Require full slot coverage for cluster to accept writes (1=yes, -1=no)",
			},
			"tls_port": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Int,
				Desc:     "Port for TLS connections (0=disabled)",
			},
			"hash_max_listpack_value": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				DataType: inputs.Int,
				Desc:     "Maximum value size in bytes for listpack encoding of hashes",
			},
			"zset_max_listpack_value": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				DataType: inputs.Int,
				Desc:     "Maximum value size in bytes for listpack encoding of sorted sets",
			},
			"lazyfree_lazy_eviction": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Use lazy freeing for evictions (1=yes, -1=no)",
			},
			"enable_module_command": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Enable module commands (1=yes, -1=no)",
			},
			"busy_reply_threshold": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				DataType: inputs.Int,
				Desc:     "Threshold in microseconds for BUSY reply",
			},
			"repl_ping_slave_period": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.DurationSecond,
				DataType: inputs.Int,
				Desc:     "Deprecated alias for `repl_ping_replica_period`",
			},
			"replica_announce_port": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Int,
				Desc:     "Port that replicas announce to master",
			},
			"slave_read_only": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Deprecated alias for `replica_read_only`",
			},
			"enable_protected_configs": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Enable protected configuration commands (1=yes, -1=no)",
			},
			"repl_diskless_sync_delay": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.DurationSecond,
				DataType: inputs.Int,
				Desc:     "Delay in seconds before starting diskless sync",
			},
			"socket_mark_id": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Int,
				Desc:     "Socket mark ID for network traffic",
			},
			"databases": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				DataType: inputs.Int,
				Desc:     "Number of databases",
			},
			"cluster_enabled": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Cluster mode enabled (1=yes, -1=no)",
			},
			"list_compress_depth": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				DataType: inputs.Int,
				Desc:     "Depth for list compression (0=disabled)",
			},
			"cluster_announce_bus_port": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Int,
				Desc:     "Cluster bus port that nodes announce to other nodes",
			},
			"replica_lazy_flush": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Use lazy flush during replica synchronization (1=yes, -1=no)",
			},
			"cluster_node_timeout": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				DataType: inputs.Int,
				Desc:     "Cluster node timeout in milliseconds",
			},
			"lfu_log_factor": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Int,
				Desc:     "Logarithmic factor for LFU (Least Frequently Used) eviction algorithm",
			},
			"hash_max_ziplist_value": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				DataType: inputs.Int,
				Desc:     "Maximum value size in bytes for ziplist encoding of hashes",
			},
			"oom_score_adj": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Adjustment for OOM (Out Of Memory) killer score (1=yes, -1=no)",
			},
			"enable_debug_command": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Enable DEBUG command (1=yes, -1=no)",
			},
			"activedefrag": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Active defragmentation enabled (1=yes, -1=no)",
			},
			"disable_thp": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Disable Transparent Huge Pages (1=yes, -1=no)",
			},
			"activerehashing": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Active rehashing enabled (1=yes, -1=no)",
			},
			"hide_user_data_from_log": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Hide user data in logs (1=yes, -1=no)",
			},
			"client_query_buffer_limit": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				DataType: inputs.Int,
				Desc:     "Client query buffer limit in bytes",
			},
			"port": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Int,
				Desc:     "TCP port Redis listens on",
			},
			"set_max_listpack_entries": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				DataType: inputs.Int,
				Desc:     "Maximum entries for listpack encoding of sets",
			},
			"active_defrag_cycle_max": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.Percent,
				DataType: inputs.Float,
				Desc:     "Maximum CPU effort percentage for active defragmentation",
			},
			"cluster_port": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Int,
				Desc:     "Cluster bus port",
			},
			"maxmemory_samples": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				DataType: inputs.Int,
				Desc:     "Number of samples for LRU/LFU eviction algorithms",
			},
			"zset_max_ziplist_entries": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				DataType: inputs.Int,
				Desc:     "Maximum entries for ziplist encoding of sorted sets",
			},
			"rdbcompression": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Compress RDB files (1=yes, -1=no)",
			},
			"timeout": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.DurationSecond,
				DataType: inputs.Int,
				Desc:     "Close idle client connections after this duration in seconds (0=disabled)",
			},
			"proto_max_bulk_len": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				DataType: inputs.Int,
				Desc:     "Maximum bulk length in bytes",
			},
			"list_max_listpack_size": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Int,
				Desc:     "Maximum size for list encoding of lists(-1:4KB/-2:8KB(default)/-3:16KB/-4:32KB/-5:64KB)",
			},
			"list_max_ziplist_size": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Int,
				Desc:     "Maximum size for ziplist encoding of lists(-1:4KB/-2:8KB(default)/-3:16KB/-4:32KB/-5:64KB)",
			},
			"cluster_link_sendbuf_limit": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				DataType: inputs.Int,
				Desc:     "Send buffer limit for cluster links in bytes",
			},
			"max_new_tls_connections_per_cycle": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				DataType: inputs.Int,
				Desc:     "Maximum new TLS connections per event loop cycle",
			},
			"replica_priority": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Int,
				Desc:     "Priority for replica to become master (lower value = higher priority)",
			},
			"cluster_slave_no_failover": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Deprecated alias for `cluster_replica_no_failover` (1=yes, -1=no)",
			},
			"no_appendfsync_on_rewrite": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Avoid fsync during AOF rewrite (1=yes, -1=no)",
			},
			"auto_aof_rewrite_min_size": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				DataType: inputs.Int,
				Desc:     "Minimum AOF size in bytes for automatic rewrite",
			},
			"slave_lazy_flush": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Deprecated alias for `replica_lazy_flush` (1=yes, -1=no)",
			},
			"tls_session_caching": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "TLS session caching enabled (1=yes, -1=no)",
			},
			"min_slaves_max_lag": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationSecond,
				Desc:     "Deprecated alias for `min_replicas_max_lag`",
			},
			"aof_load_truncated": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Load truncated AOF files (1=yes, -1=no)",
			},
			"lazyfree_lazy_user_del": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Use lazy freeing for user DEL commands (1=yes, -1=no)",
			},
			"latency_tracking": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Latency tracking enabled (1=yes, -1=no)",
			},
			"replica_ignore_maxmemory": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Int,
				Desc:     "Replicas ignore maxmemory setting (1=yes, -1=no)",
			},
			"rdb_del_sync_files": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Delete RDB files after successful synchronization (1=yes, -1=no)",
			},
			"tls_session_cache_timeout": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.DurationSecond,
				DataType: inputs.Int,
				Desc:     "Timeout in seconds for TLS session cache entries",
			},
			"tls_cluster": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Use TLS for cluster bus communication (1=yes, -1=no)",
			},
			"protected_mode": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Protected mode enabled (1=yes, -1=no)",
			},
			"lazyfree_lazy_user_flush": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Use lazy freeing for FLUSH commands (1=yes, -1=no)",
			},
			"tls_auth_clients": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Require client authentication for TLS connections (1=yes, -1=no)",
			},
			"slave_ignore_maxmemory": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Deprecated alias for `replica_ignore_maxmemory`",
			},
			"replica_serve_stale_data": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				DataType: inputs.Bool,
				Desc:     "Replicas serve stale data when master is down (1=yes, -1=no)",
			},
		},
	}
}
