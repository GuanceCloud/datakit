// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func Test_infoMeasurement_getErrorPoints(t *testing.T) {
	type fields struct {
		cli         *redis.Client
		name        string
		tags        map[string]string
		fields      map[string]interface{}
		resData     map[string]interface{}
		election    bool
		lastCollect *redisCPUUsage
	}
	type args struct {
		k string
		v string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			name: "ok 0",
			fields: fields{
				cli:         &redis.Client{},
				name:        "redis_indo",
				tags:        map[string]string{"foo": "bar"},
				fields:      map[string]interface{}{},
				resData:     map[string]interface{}{},
				election:    true,
				lastCollect: &redisCPUUsage{},
			},
			args: args{
				k: "errorstat_ERR",
				v: "count=188",
			},
			want: []string{
				"redis_indo,error_type=ERR,foo=bar errorstat=188i",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &infoMeasurement{
				cli:         tt.fields.cli,
				name:        tt.fields.name,
				tags:        tt.fields.tags,
				lastCollect: tt.fields.lastCollect,
			}
			got := m.getErrorPoints(tt.args.k, tt.args.v)

			gotStr := []string{}
			for _, v := range got {
				s := v.LineProto()
				s = s[:strings.LastIndex(s, " ")]
				gotStr = append(gotStr, s)
			}
			sort.Strings(gotStr)

			assert.Equal(t, gotStr, tt.want)
		})
	}
}

func Test_infoMeasurement_getLatencyPoints(t *testing.T) {
	type fields struct {
		cli                *redis.Client
		name               string
		tags               map[string]string
		fields             map[string]interface{}
		resData            map[string]interface{}
		election           bool
		lastCollect        *redisCPUUsage
		latencyPercentiles bool
	}
	type args struct {
		k string
		v string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			name: "ok 0",
			fields: fields{
				cli:                &redis.Client{},
				name:               "redis_indo",
				tags:               map[string]string{"foo": "bar"},
				fields:             map[string]interface{}{},
				resData:            map[string]interface{}{},
				election:           true,
				lastCollect:        &redisCPUUsage{},
				latencyPercentiles: true,
			},
			args: args{
				k: "latency_percentiles_usec_client|list",
				v: "p50=23.039,p99=70.143,p99.9=70.143",
			},
			want: []string{
				"redis_indo,command_type=client|list,foo=bar,quantile=0.5 latency_percentiles_usec=23.039",
				"redis_indo,command_type=client|list,foo=bar,quantile=0.99 latency_percentiles_usec=70.143",
				"redis_indo,command_type=client|list,foo=bar,quantile=0.999 latency_percentiles_usec=70.143",
			},
		},
		{
			name: "not command stats",
			fields: fields{
				cli:                &redis.Client{},
				name:               "redis_indo",
				tags:               map[string]string{"foo": "bar"},
				fields:             map[string]interface{}{},
				resData:            map[string]interface{}{},
				election:           true,
				lastCollect:        &redisCPUUsage{},
				latencyPercentiles: false,
			},
			args: args{
				k: "latency_percentiles_usec_client|list",
				v: "p50=23.039,p99=70.143,p99.9=70.143",
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &infoMeasurement{
				cli:                tt.fields.cli,
				name:               tt.fields.name,
				tags:               tt.fields.tags,
				lastCollect:        tt.fields.lastCollect,
				latencyPercentiles: tt.fields.latencyPercentiles,
			}
			got := m.getLatencyPoints(tt.args.k, tt.args.v)

			gotStr := []string{}
			for _, v := range got {
				s := v.LineProto()
				s = s[:strings.LastIndex(s, " ")]
				gotStr = append(gotStr, s)
			}
			sort.Strings(gotStr)

			assert.Equal(t, gotStr, tt.want)
		})
	}
}

func Test_getLatencyTagField(t *testing.T) {
	type args struct {
		v string
	}
	tests := []struct {
		name string
		args args
		want map[string]float64
	}{
		{
			name: "ok 0",
			args: args{
				v: "p50=23.039,p99=70.143,p99.9=70.143",
			},
			want: map[string]float64{"0.5": 23.039, "0.99": 70.143, "0.999": 70.143},
		},
		{
			name: "ok 1",
			args: args{
				v: "50=23.039,99=70.143,99.9=70.143",
			},
			want: map[string]float64{"0.5": 23.039, "0.99": 70.143, "0.999": 70.143},
		},
		{
			name: "have dome error",
			args: args{
				v: "X50=23.039,y99=70.143,99.9=70.143",
			},
			want: map[string]float64{"0.999": 70.143},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getLatencyTagField(tt.args.v)

			assert.Equal(t, got, tt.want)
		})
	}
}

func Test_getQuantile(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "ok 0",
			args: args{
				s: "50",
			},
			want:    "0.5",
			wantErr: false,
		},
		{
			name: "ok 1",
			args: args{
				s: "99",
			},
			want:    "0.99",
			wantErr: false,
		},
		{
			name: "ok 2",
			args: args{
				s: "99.9",
			},
			want:    "0.999",
			wantErr: false,
		},
		{
			name: "ok 3",
			args: args{
				s: ".0001",
			},
			want:    "0.000001",
			wantErr: false,
		},
		{
			name: "ok 4",
			args: args{
				s: "1000",
			},
			want:    "10",
			wantErr: false,
		},
		{
			name: "error max 6 decimal places",
			args: args{
				s: ".00001",
			},
			want:    "0.0000001",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getQuantile(tt.args.s)

			if (err != nil) != tt.wantErr {
				t.Errorf("getQuantile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err != nil) && tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("getQuantile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_infoMeasurement_parseInfoData(t *testing.T) {
	type fields struct {
		cli                *redis.Client
		name               string
		tags               map[string]string
		fields             map[string]interface{}
		resData            map[string]interface{}
		lastCollect        *redisCPUUsage
		latencyPercentiles bool
	}
	type args struct {
		info   string
		info02 string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "ok latencyPercentiles=false",
			fields: fields{
				cli:         &redis.Client{},
				name:        "redis_indo",
				tags:        map[string]string{"foo": "bar"},
				fields:      map[string]interface{}{},
				resData:     map[string]interface{}{},
				lastCollect: &redisCPUUsage{},
				// latencyPercentiles: true,
			},
			args: args{
				info:   mockInfo,
				info02: mockInfo,
			},
			want: []string{
				"redis_indo,error_type=WRONGPASS,foo=bar,redis_version=7.0.13 errorstat=8i",
				"redis_indo,foo=bar,redis_version=7.0.13 active_defrag_hits=0,active_defrag_key_hits=0,active_defrag_key_misses=0,active_defrag_misses=0,active_defrag_running=0,allocator_active=2080768,allocator_allocated=1738944,allocator_frag_bytes=341824,allocator_frag_ratio=1.2,allocator_resident=4939776,allocator_rss_bytes=2859008,allocator_rss_ratio=2.37,aof_base_size=89,aof_buffer_length=0,aof_current_rewrite_time_sec=-1,aof_current_size=89,aof_delayed_fsync=0,aof_enabled=1,aof_last_cow_size=0,aof_last_rewrite_time_sec=-1,aof_pending_bio_fsync=0,aof_pending_rewrite=0,aof_rewrite_in_progress=0,aof_rewrite_scheduled=0,aof_rewrites=0,arch_bits=64,async_loading=0,blocked_clients=0,client_recent_max_input_buffer=8,client_recent_max_output_buffer=0,clients_in_timeout_table=0,cluster_connections=0,cluster_enabled=0,configured_hz=10,connected_clients=1,connected_slaves=0,current_active_defrag_time=0,current_cow_peak=0,current_cow_size=0,current_cow_size_age=0,current_eviction_exceeded_time=0,current_fork_perc=0,current_save_keys_processed=0,current_save_keys_total=0,dump_payload_sanitizations=0,evicted_clients=0,evicted_keys=0,expire_cycle_cpu_milliseconds=29,expired_keys=0,expired_stale_perc=0,expired_time_cap_reached_count=0,hz=10,info_latency_ms=3.73,instantaneous_input_kbps=0,instantaneous_input_repl_kbps=0,instantaneous_ops_per_sec=0,instantaneous_output_kbps=0,instantaneous_output_repl_kbps=0,io_threaded_reads_processed=0,io_threaded_writes_processed=0,io_threads_active=0,keyspace_hits=0,keyspace_misses=0,latest_fork_usec=0,lazyfree_pending_objects=0,lazyfreed_objects=0,loading=0,lru_clock=4153786,master_repl_offset=0,maxclients=10000,maxmemory=0,mem_aof_buffer=8,mem_clients_normal=1800,mem_clients_slaves=0,mem_cluster_links=0,mem_fragmentation_bytes=5813192,mem_fragmentation_ratio=6.46,mem_not_counted_for_evict=8,mem_replication_backlog=0,mem_total_replication_buffers=0,migrate_cached_sockets=0,module_fork_in_progress=0,module_fork_last_cow_size=0,pubsub_channels=0,pubsub_patterns=0,pubsubshard_channels=0,rdb_bgsave_in_progress=0,rdb_changes_since_last_save=0,rdb_current_bgsave_time_sec=-1,rdb_last_bgsave_time_sec=-1,rdb_last_cow_size=0,rdb_last_load_keys_expired=0,rdb_last_load_keys_loaded=0,rdb_last_save_time=1698651554,rdb_saves=0,rejected_connections=0,repl_backlog_active=0,repl_backlog_first_byte_offset=0,repl_backlog_histlen=0,repl_backlog_size=1048576,rss_overhead_bytes=1937408,rss_overhead_ratio=1.39,second_repl_offset=-1,server_time_usec=1698652602514848,slave_expires_tracked_keys=0,sync_full=0,sync_partial_err=0,sync_partial_ok=0,tcp_port=6379,total_active_defrag_time=0,total_commands_processed=211,total_connections_received=10,total_error_replies=8,total_eviction_exceeded_time=0,total_forks=0,total_net_input_bytes=6445,total_net_output_bytes=407193,total_net_repl_input_bytes=0,total_net_repl_output_bytes=0,total_reads_processed=221,total_system_memory=16826019840,total_writes_processed=213,tracking_clients=0,tracking_total_items=0,tracking_total_keys=0,tracking_total_prefixes=0,unexpected_error_replies=0,uptime_in_days=0,uptime_in_seconds=1048,used_cpu_sys=1.300455,used_cpu_sys_children=0.004007,used_cpu_sys_main_thread=1.296392,used_cpu_sys_percent=0,used_cpu_user=1.717573,used_cpu_user_children=0.007766,used_cpu_user_main_thread=1.714453,used_cpu_user_percent=0,used_memory=1086408,used_memory_dataset=222144,used_memory_dataset_perc=99.11,used_memory_lua=31744,used_memory_overhead=864264,used_memory_peak=1115248,used_memory_peak_perc=97.41,used_memory_rss=6877184,used_memory_scripts=184,used_memory_startup=862272",
			},
			wantErr: false,
		},
		{
			name: "ok latencyPercentiles=true",
			fields: fields{
				cli:                &redis.Client{},
				name:               "redis_indo",
				tags:               map[string]string{"foo": "bar"},
				fields:             map[string]interface{}{},
				resData:            map[string]interface{}{},
				lastCollect:        &redisCPUUsage{},
				latencyPercentiles: true,
			},
			args: args{
				info:   mockInfo,
				info02: mockInfo,
			},
			want: []string{
				"redis_indo,command_type=acl|setuser,foo=bar,quantile=0.5,redis_version=7.0.13 latency_percentiles_usec=78.335",
				"redis_indo,command_type=acl|setuser,foo=bar,quantile=0.99,redis_version=7.0.13 latency_percentiles_usec=103.423",
				"redis_indo,command_type=acl|setuser,foo=bar,quantile=0.999,redis_version=7.0.13 latency_percentiles_usec=103.423",
				"redis_indo,command_type=auth,foo=bar,quantile=0.5,redis_version=7.0.13 latency_percentiles_usec=51.199",
				"redis_indo,command_type=auth,foo=bar,quantile=0.99,redis_version=7.0.13 latency_percentiles_usec=114.175",
				"redis_indo,command_type=auth,foo=bar,quantile=0.999,redis_version=7.0.13 latency_percentiles_usec=114.175",
				"redis_indo,command_type=client|list,foo=bar,quantile=0.5,redis_version=7.0.13 latency_percentiles_usec=35.071",
				"redis_indo,command_type=client|list,foo=bar,quantile=0.99,redis_version=7.0.13 latency_percentiles_usec=88.063",
				"redis_indo,command_type=client|list,foo=bar,quantile=0.999,redis_version=7.0.13 latency_percentiles_usec=88.063",
				"redis_indo,command_type=command|docs,foo=bar,quantile=0.5,redis_version=7.0.13 latency_percentiles_usec=1835.007",
				"redis_indo,command_type=command|docs,foo=bar,quantile=0.99,redis_version=7.0.13 latency_percentiles_usec=1835.007",
				"redis_indo,command_type=command|docs,foo=bar,quantile=0.999,redis_version=7.0.13 latency_percentiles_usec=1835.007",
				"redis_indo,command_type=info,foo=bar,quantile=0.5,redis_version=7.0.13 latency_percentiles_usec=97.279",
				"redis_indo,command_type=info,foo=bar,quantile=0.99,redis_version=7.0.13 latency_percentiles_usec=581.631",
				"redis_indo,command_type=info,foo=bar,quantile=0.999,redis_version=7.0.13 latency_percentiles_usec=708.607",
				"redis_indo,command_type=latency|latest,foo=bar,quantile=0.5,redis_version=7.0.13 latency_percentiles_usec=4.015",
				"redis_indo,command_type=latency|latest,foo=bar,quantile=0.99,redis_version=7.0.13 latency_percentiles_usec=30.079",
				"redis_indo,command_type=latency|latest,foo=bar,quantile=0.999,redis_version=7.0.13 latency_percentiles_usec=30.079",
				"redis_indo,command_type=ping,foo=bar,quantile=0.5,redis_version=7.0.13 latency_percentiles_usec=2.007",
				"redis_indo,command_type=ping,foo=bar,quantile=0.99,redis_version=7.0.13 latency_percentiles_usec=2.007",
				"redis_indo,command_type=ping,foo=bar,quantile=0.999,redis_version=7.0.13 latency_percentiles_usec=2.007",
				"redis_indo,command_type=slowlog|get,foo=bar,quantile=0.5,redis_version=7.0.13 latency_percentiles_usec=5.023",
				"redis_indo,command_type=slowlog|get,foo=bar,quantile=0.99,redis_version=7.0.13 latency_percentiles_usec=24.063",
				"redis_indo,command_type=slowlog|get,foo=bar,quantile=0.999,redis_version=7.0.13 latency_percentiles_usec=24.063",
				"redis_indo,error_type=WRONGPASS,foo=bar,redis_version=7.0.13 errorstat=8i",
				"redis_indo,foo=bar,redis_version=7.0.13 active_defrag_hits=0,active_defrag_key_hits=0,active_defrag_key_misses=0,active_defrag_misses=0,active_defrag_running=0,allocator_active=2080768,allocator_allocated=1738944,allocator_frag_bytes=341824,allocator_frag_ratio=1.2,allocator_resident=4939776,allocator_rss_bytes=2859008,allocator_rss_ratio=2.37,aof_base_size=89,aof_buffer_length=0,aof_current_rewrite_time_sec=-1,aof_current_size=89,aof_delayed_fsync=0,aof_enabled=1,aof_last_cow_size=0,aof_last_rewrite_time_sec=-1,aof_pending_bio_fsync=0,aof_pending_rewrite=0,aof_rewrite_in_progress=0,aof_rewrite_scheduled=0,aof_rewrites=0,arch_bits=64,async_loading=0,blocked_clients=0,client_recent_max_input_buffer=8,client_recent_max_output_buffer=0,clients_in_timeout_table=0,cluster_connections=0,cluster_enabled=0,configured_hz=10,connected_clients=1,connected_slaves=0,current_active_defrag_time=0,current_cow_peak=0,current_cow_size=0,current_cow_size_age=0,current_eviction_exceeded_time=0,current_fork_perc=0,current_save_keys_processed=0,current_save_keys_total=0,dump_payload_sanitizations=0,evicted_clients=0,evicted_keys=0,expire_cycle_cpu_milliseconds=29,expired_keys=0,expired_stale_perc=0,expired_time_cap_reached_count=0,hz=10,info_latency_ms=3.73,instantaneous_input_kbps=0,instantaneous_input_repl_kbps=0,instantaneous_ops_per_sec=0,instantaneous_output_kbps=0,instantaneous_output_repl_kbps=0,io_threaded_reads_processed=0,io_threaded_writes_processed=0,io_threads_active=0,keyspace_hits=0,keyspace_misses=0,latest_fork_usec=0,lazyfree_pending_objects=0,lazyfreed_objects=0,loading=0,lru_clock=4153786,master_repl_offset=0,maxclients=10000,maxmemory=0,mem_aof_buffer=8,mem_clients_normal=1800,mem_clients_slaves=0,mem_cluster_links=0,mem_fragmentation_bytes=5813192,mem_fragmentation_ratio=6.46,mem_not_counted_for_evict=8,mem_replication_backlog=0,mem_total_replication_buffers=0,migrate_cached_sockets=0,module_fork_in_progress=0,module_fork_last_cow_size=0,pubsub_channels=0,pubsub_patterns=0,pubsubshard_channels=0,rdb_bgsave_in_progress=0,rdb_changes_since_last_save=0,rdb_current_bgsave_time_sec=-1,rdb_last_bgsave_time_sec=-1,rdb_last_cow_size=0,rdb_last_load_keys_expired=0,rdb_last_load_keys_loaded=0,rdb_last_save_time=1698651554,rdb_saves=0,rejected_connections=0,repl_backlog_active=0,repl_backlog_first_byte_offset=0,repl_backlog_histlen=0,repl_backlog_size=1048576,rss_overhead_bytes=1937408,rss_overhead_ratio=1.39,second_repl_offset=-1,server_time_usec=1698652602514848,slave_expires_tracked_keys=0,sync_full=0,sync_partial_err=0,sync_partial_ok=0,tcp_port=6379,total_active_defrag_time=0,total_commands_processed=211,total_connections_received=10,total_error_replies=8,total_eviction_exceeded_time=0,total_forks=0,total_net_input_bytes=6445,total_net_output_bytes=407193,total_net_repl_input_bytes=0,total_net_repl_output_bytes=0,total_reads_processed=221,total_system_memory=16826019840,total_writes_processed=213,tracking_clients=0,tracking_total_items=0,tracking_total_keys=0,tracking_total_prefixes=0,unexpected_error_replies=0,uptime_in_days=0,uptime_in_seconds=1048,used_cpu_sys=1.300455,used_cpu_sys_children=0.004007,used_cpu_sys_main_thread=1.296392,used_cpu_sys_percent=0,used_cpu_user=1.717573,used_cpu_user_children=0.007766,used_cpu_user_main_thread=1.714453,used_cpu_user_percent=0,used_memory=1086408,used_memory_dataset=222144,used_memory_dataset_perc=99.11,used_memory_lua=31744,used_memory_overhead=864264,used_memory_peak=1115248,used_memory_peak_perc=97.41,used_memory_rss=6877184,used_memory_scripts=184,used_memory_startup=862272",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &infoMeasurement{
				cli:                tt.fields.cli,
				name:               tt.fields.name,
				tags:               tt.fields.tags,
				lastCollect:        tt.fields.lastCollect,
				latencyPercentiles: tt.fields.latencyPercentiles,
			}

			// loop 1
			start := time.Now()
			aa := start.Unix()
			start2 := time.Unix(1698724562, 0)
			bb := start2.Unix()
			_, _ = aa, bb

			info := tt.args.info
			elapsed := time.Duration(3728281)

			nextTS := start.Add(elapsed / 2)

			latencyMs := Round(float64(elapsed)/float64(time.Millisecond), 2)

			_, err := m.parseInfoData(info, latencyMs, nextTS)
			if (err != nil) != tt.wantErr {
				t.Errorf("infoMeasurement.parseInfoData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// loop 2
			time.Sleep(time.Second)

			start = time.Now()

			info = tt.args.info02
			elapsed = time.Duration(3728281)

			nextTS = start.Add(elapsed / 2)

			latencyMs = Round(float64(elapsed)/float64(time.Millisecond), 2)

			got, err := m.parseInfoData(info, latencyMs, nextTS)
			if (err != nil) != tt.wantErr {
				t.Errorf("infoMeasurement.parseInfoData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			gotStr := []string{}
			for _, v := range got {
				s := v.LineProto()
				s = s[:strings.LastIndex(s, " ")]
				gotStr = append(gotStr, s)
			}
			sort.Strings(gotStr)

			assert.Equal(t, gotStr, tt.want)
		})
	}
}

var mockInfo = `# Server
redis_version:7.0.13
redis_git_sha1:00000000
redis_git_dirty:0
redis_build_id:e9f5dfc882060196
redis_mode:standalone
os:Linux 3.10.0-957.el7.x86_64 x86_64
arch_bits:64
monotonic_clock:POSIX clock_gettime
multiplexing_api:epoll
atomicvar_api:c11-builtin
gcc_version:12.2.0
process_id:1
process_supervised:no
run_id:85339d49fa59b92c47a3da041dd85b104af97f27
tcp_port:6379
server_time_usec:1698652602514848
uptime_in_seconds:1048
uptime_in_days:0
hz:10
configured_hz:10
lru_clock:4153786
executable:/data/redis-server
config_file:/etc/redis/redis.conf
io_threads_active:0

# Clients
connected_clients:1
cluster_connections:0
maxclients:10000
client_recent_max_input_buffer:8
client_recent_max_output_buffer:0
blocked_clients:0
tracking_clients:0
clients_in_timeout_table:0

# Memory
used_memory:1086408
used_memory_human:1.04M
used_memory_rss:6877184
used_memory_rss_human:6.56M
used_memory_peak:1115248
used_memory_peak_human:1.06M
used_memory_peak_perc:97.41%
used_memory_overhead:864264
used_memory_startup:862272
used_memory_dataset:222144
used_memory_dataset_perc:99.11%
allocator_allocated:1738944
allocator_active:2080768
allocator_resident:4939776
total_system_memory:16826019840
total_system_memory_human:15.67G
used_memory_lua:31744
used_memory_vm_eval:31744
used_memory_lua_human:31.00K
used_memory_scripts_eval:0
number_of_cached_scripts:0
number_of_functions:0
number_of_libraries:0
used_memory_vm_functions:32768
used_memory_vm_total:64512
used_memory_vm_total_human:63.00K
used_memory_functions:184
used_memory_scripts:184
used_memory_scripts_human:184B
maxmemory:0
maxmemory_human:0B
maxmemory_policy:noeviction
allocator_frag_ratio:1.20
allocator_frag_bytes:341824
allocator_rss_ratio:2.37
allocator_rss_bytes:2859008
rss_overhead_ratio:1.39
rss_overhead_bytes:1937408
mem_fragmentation_ratio:6.46
mem_fragmentation_bytes:5813192
mem_not_counted_for_evict:8
mem_replication_backlog:0
mem_total_replication_buffers:0
mem_clients_slaves:0
mem_clients_normal:1800
mem_cluster_links:0
mem_aof_buffer:8
mem_allocator:jemalloc-5.2.1
active_defrag_running:0
lazyfree_pending_objects:0
lazyfreed_objects:0

# Persistence
loading:0
async_loading:0
current_cow_peak:0
current_cow_size:0
current_cow_size_age:0
current_fork_perc:0.00
current_save_keys_processed:0
current_save_keys_total:0
rdb_changes_since_last_save:0
rdb_bgsave_in_progress:0
rdb_last_save_time:1698651554
rdb_last_bgsave_status:ok
rdb_last_bgsave_time_sec:-1
rdb_current_bgsave_time_sec:-1
rdb_saves:0
rdb_last_cow_size:0
rdb_last_load_keys_expired:0
rdb_last_load_keys_loaded:0
aof_enabled:1
aof_rewrite_in_progress:0
aof_rewrite_scheduled:0
aof_last_rewrite_time_sec:-1
aof_current_rewrite_time_sec:-1
aof_last_bgrewrite_status:ok
aof_rewrites:0
aof_rewrites_consecutive_failures:0
aof_last_write_status:ok
aof_last_cow_size:0
module_fork_in_progress:0
module_fork_last_cow_size:0
aof_current_size:89
aof_base_size:89
aof_pending_rewrite:0
aof_buffer_length:0
aof_pending_bio_fsync:0
aof_delayed_fsync:0

# Stats
total_connections_received:10
total_commands_processed:211
instantaneous_ops_per_sec:0
total_net_input_bytes:6445
total_net_output_bytes:407193
total_net_repl_input_bytes:0
total_net_repl_output_bytes:0
instantaneous_input_kbps:0.00
instantaneous_output_kbps:0.00
instantaneous_input_repl_kbps:0.00
instantaneous_output_repl_kbps:0.00
rejected_connections:0
sync_full:0
sync_partial_ok:0
sync_partial_err:0
expired_keys:0
expired_stale_perc:0.00
expired_time_cap_reached_count:0
expire_cycle_cpu_milliseconds:29
evicted_keys:0
evicted_clients:0
total_eviction_exceeded_time:0
current_eviction_exceeded_time:0
keyspace_hits:0
keyspace_misses:0
pubsub_channels:0
pubsub_patterns:0
pubsubshard_channels:0
latest_fork_usec:0
total_forks:0
migrate_cached_sockets:0
slave_expires_tracked_keys:0
active_defrag_hits:0
active_defrag_misses:0
active_defrag_key_hits:0
active_defrag_key_misses:0
total_active_defrag_time:0
current_active_defrag_time:0
tracking_total_keys:0
tracking_total_items:0
tracking_total_prefixes:0
unexpected_error_replies:0
total_error_replies:8
dump_payload_sanitizations:0
total_reads_processed:221
total_writes_processed:213
io_threaded_reads_processed:0
io_threaded_writes_processed:0
reply_buffer_shrinks:28
reply_buffer_expands:26

# Replication
role:master
connected_slaves:0
master_failover_state:no-failover
master_replid:b853cdaf3b3ea32e2c4064f4fe6bb75ec9278ce7
master_replid2:0000000000000000000000000000000000000000
master_repl_offset:0
second_repl_offset:-1
repl_backlog_active:0
repl_backlog_size:1048576
repl_backlog_first_byte_offset:0
repl_backlog_histlen:0

# CPU
used_cpu_sys:1.300455
used_cpu_user:1.717573
used_cpu_sys_children:0.004007
used_cpu_user_children:0.007766
used_cpu_sys_main_thread:1.296392
used_cpu_user_main_thread:1.714453

# Modules

# Commandstats
cmdstat_ping:calls=1,usec=2,usec_per_call=2.00,rejected_calls=0,failed_calls=0
cmdstat_auth:calls=9,usec=549,usec_per_call=61.00,rejected_calls=0,failed_calls=8
cmdstat_latency|latest:calls=28,usec=173,usec_per_call=6.18,rejected_calls=0,failed_calls=0
cmdstat_command|docs:calls=1,usec=1829,usec_per_call=1829.00,rejected_calls=0,failed_calls=0
cmdstat_client|list:calls=28,usec=1035,usec_per_call=36.96,rejected_calls=0,failed_calls=0
cmdstat_info:calls=113,usec=19830,usec_per_call=175.49,rejected_calls=0,failed_calls=0
cmdstat_slowlog|get:calls=28,usec=169,usec_per_call=6.04,rejected_calls=0,failed_calls=0
cmdstat_acl|setuser:calls=3,usec=226,usec_per_call=75.33,rejected_calls=0,failed_calls=0

# Errorstats
errorstat_WRONGPASS:count=8

# Latencystats
latency_percentiles_usec_ping:p50=2.007,p99=2.007,p99.9=2.007
latency_percentiles_usec_auth:p50=51.199,p99=114.175,p99.9=114.175
latency_percentiles_usec_latency|latest:p50=4.015,p99=30.079,p99.9=30.079
latency_percentiles_usec_command|docs:p50=1835.007,p99=1835.007,p99.9=1835.007
latency_percentiles_usec_client|list:p50=35.071,p99=88.063,p99.9=88.063
latency_percentiles_usec_info:p50=97.279,p99=581.631,p99.9=708.607
latency_percentiles_usec_slowlog|get:p50=5.023,p99=24.063,p99.9=24.063
latency_percentiles_usec_acl|setuser:p50=78.335,p99=103.423,p99.9=103.423

# Cluster
cluster_enabled:0

# Keyspace
`

var mockInfo01 = `# Server
redis_version:7.0.13
redis_git_sha1:00000000
redis_git_dirty:0
redis_build_id:e9f5dfc882060196
redis_mode:standalone
os:Linux 3.10.0-957.el7.x86_64 x86_64
arch_bits:64
monotonic_clock:POSIX clock_gettime
multiplexing_api:epoll
atomicvar_api:c11-builtin
gcc_version:12.2.0
process_id:1
process_supervised:no
run_id:85339d49fa59b92c47a3da041dd85b104af97f27
tcp_port:6379
server_time_usec:1698654353760906
uptime_in_seconds:2799
uptime_in_days:0
hz:10
configured_hz:10
lru_clock:4155537
executable:/data/redis-server
config_file:/etc/redis/redis.conf
io_threads_active:0

# Clients
connected_clients:2
cluster_connections:0
maxclients:10000
client_recent_max_input_buffer:20480
client_recent_max_output_buffer:0
blocked_clients:0
tracking_clients:0
clients_in_timeout_table:0

# Memory
used_memory:1088608
used_memory_human:1.04M
used_memory_rss:6877184
used_memory_rss_human:6.56M
used_memory_peak:1266152
used_memory_peak_human:1.21M
used_memory_peak_perc:85.98%
used_memory_overhead:866064
used_memory_startup:862272
used_memory_dataset:222544
used_memory_dataset_perc:98.32%
allocator_allocated:1673496
allocator_active:2080768
allocator_resident:4939776
total_system_memory:16826019840
total_system_memory_human:15.67G
used_memory_lua:31744
used_memory_vm_eval:31744
used_memory_lua_human:31.00K
used_memory_scripts_eval:0
number_of_cached_scripts:0
number_of_functions:0
number_of_libraries:0
used_memory_vm_functions:32768
used_memory_vm_total:64512
used_memory_vm_total_human:63.00K
used_memory_functions:184
used_memory_scripts:184
used_memory_scripts_human:184B
maxmemory:0
maxmemory_human:0B
maxmemory_policy:noeviction
allocator_frag_ratio:1.24
allocator_frag_bytes:407272
allocator_rss_ratio:2.37
allocator_rss_bytes:2859008
rss_overhead_ratio:1.39
rss_overhead_bytes:1937408
mem_fragmentation_ratio:6.45
mem_fragmentation_bytes:5810992
mem_not_counted_for_evict:8
mem_replication_backlog:0
mem_total_replication_buffers:0
mem_clients_slaves:0
mem_clients_normal:3600
mem_cluster_links:0
mem_aof_buffer:8
mem_allocator:jemalloc-5.2.1
active_defrag_running:0
lazyfree_pending_objects:0
lazyfreed_objects:0

# Persistence
loading:0
async_loading:0
current_cow_peak:0
current_cow_size:0
current_cow_size_age:0
current_fork_perc:0.00
current_save_keys_processed:0
current_save_keys_total:0
rdb_changes_since_last_save:0
rdb_bgsave_in_progress:0
rdb_last_save_time:1698651554
rdb_last_bgsave_status:ok
rdb_last_bgsave_time_sec:-1
rdb_current_bgsave_time_sec:-1
rdb_saves:0
rdb_last_cow_size:0
rdb_last_load_keys_expired:0
rdb_last_load_keys_loaded:0
aof_enabled:1
aof_rewrite_in_progress:0
aof_rewrite_scheduled:0
aof_last_rewrite_time_sec:-1
aof_current_rewrite_time_sec:-1
aof_last_bgrewrite_status:ok
aof_rewrites:0
aof_rewrites_consecutive_failures:0
aof_last_write_status:ok
aof_last_cow_size:0
module_fork_in_progress:0
module_fork_last_cow_size:0
aof_current_size:89
aof_base_size:89
aof_pending_rewrite:0
aof_buffer_length:0
aof_pending_bio_fsync:0
aof_delayed_fsync:0

# Stats
total_connections_received:12
total_commands_processed:214
instantaneous_ops_per_sec:0
total_net_input_bytes:6522
total_net_output_bytes:756843
total_net_repl_input_bytes:0
total_net_repl_output_bytes:0
instantaneous_input_kbps:0.00
instantaneous_output_kbps:0.00
instantaneous_input_repl_kbps:0.00
instantaneous_output_repl_kbps:0.00
rejected_connections:0
sync_full:0
sync_partial_ok:0
sync_partial_err:0
expired_keys:0
expired_stale_perc:0.00
expired_time_cap_reached_count:0
expire_cycle_cpu_milliseconds:80
evicted_keys:0
evicted_clients:0
total_eviction_exceeded_time:0
current_eviction_exceeded_time:0
keyspace_hits:0
keyspace_misses:0
pubsub_channels:0
pubsub_patterns:0
pubsubshard_channels:0
latest_fork_usec:0
total_forks:0
migrate_cached_sockets:0
slave_expires_tracked_keys:0
active_defrag_hits:0
active_defrag_misses:0
active_defrag_key_hits:0
active_defrag_key_misses:0
total_active_defrag_time:0
current_active_defrag_time:0
tracking_total_keys:0
tracking_total_items:0
tracking_total_prefixes:0
unexpected_error_replies:0
total_error_replies:8
dump_payload_sanitizations:0
total_reads_processed:225
total_writes_processed:220
io_threaded_reads_processed:0
io_threaded_writes_processed:0
reply_buffer_shrinks:31
reply_buffer_expands:27

# Replication
role:master
connected_slaves:0
master_failover_state:no-failover
master_replid:b853cdaf3b3ea32e2c4064f4fe6bb75ec9278ce7
master_replid2:0000000000000000000000000000000000000000
master_repl_offset:0
second_repl_offset:-1
repl_backlog_active:0
repl_backlog_size:1048576
repl_backlog_first_byte_offset:0
repl_backlog_histlen:0

# CPU
used_cpu_sys:3.359672
used_cpu_user:4.367579
used_cpu_sys_children:0.004007
used_cpu_user_children:0.007766
used_cpu_sys_main_thread:3.355627
used_cpu_user_main_thread:4.364442

# Modules

# Commandstats
cmdstat_ping:calls=1,usec=2,usec_per_call=2.00,rejected_calls=0,failed_calls=0
cmdstat_auth:calls=9,usec=549,usec_per_call=61.00,rejected_calls=0,failed_calls=8
cmdstat_latency|latest:calls=28,usec=173,usec_per_call=6.18,rejected_calls=0,failed_calls=0
cmdstat_command|docs:calls=3,usec=6016,usec_per_call=2005.33,rejected_calls=0,failed_calls=0
cmdstat_client|list:calls=28,usec=1035,usec_per_call=36.96,rejected_calls=0,failed_calls=0
cmdstat_info:calls=114,usec=20364,usec_per_call=178.63,rejected_calls=0,failed_calls=0
cmdstat_slowlog|get:calls=28,usec=169,usec_per_call=6.04,rejected_calls=0,failed_calls=0
cmdstat_acl|setuser:calls=3,usec=226,usec_per_call=75.33,rejected_calls=0,failed_calls=0

# Errorstats
errorstat_WRONGPASS:count=8

# Latencystats
latency_percentiles_usec_ping:p50=2.007,p99=2.007,p99.9=2.007
latency_percentiles_usec_auth:p50=51.199,p99=114.175,p99.9=114.175
latency_percentiles_usec_latency|latest:p50=4.015,p99=30.079,p99.9=30.079
latency_percentiles_usec_command|docs:p50=1835.007,p99=2473.983,p99.9=2473.983
latency_percentiles_usec_client|list:p50=35.071,p99=88.063,p99.9=88.063
latency_percentiles_usec_info:p50=97.279,p99=581.631,p99.9=708.607
latency_percentiles_usec_slowlog|get:p50=5.023,p99=24.063,p99.9=24.063
latency_percentiles_usec_acl|setuser:p50=78.335,p99=103.423,p99.9=103.423

# Cluster
cluster_enabled:0

# Keyspace
`
