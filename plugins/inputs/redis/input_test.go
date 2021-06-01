package redis

import (
	"fmt"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
	"testing"
	// "gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var redisInfo = `
# Server
redis_version:5.0.4
redis_git_sha1:00000000
redis_git_dirty:0
redis_build_id:b0d03e9600ca5411
redis_mode:standalone
os:Linux 4.9.125-linuxkit x86_64
arch_bits:64
multiplexing_api:epoll
atomicvar_api:atomic-builtin
gcc_version:8.3.0
process_id:1
run_id:4c879f8985bfc246b27f142274e03b6ff6dbbebe
tcp_port:6379
uptime_in_seconds:629932
uptime_in_days:7
hz:10
configured_hz:10
lru_clock:11346140
executable:/data/redis-server
config_file:

# Clients
connected_clients:2
client_recent_max_input_buffer:2
client_recent_max_output_buffer:0
blocked_clients:0

# Memory
used_memory:875152
used_memory_human:854.64K
used_memory_rss:3047424
used_memory_rss_human:2.91M
used_memory_peak:956792
used_memory_peak_human:934.37K
used_memory_peak_perc:91.47%
used_memory_overhead:857760
used_memory_startup:791000
used_memory_dataset:17392
used_memory_dataset_perc:20.67%
allocator_allocated:871552
allocator_active:1105920
allocator_resident:4263936
total_system_memory:2095869952
total_system_memory_human:1.95G
used_memory_lua:37888
used_memory_lua_human:37.00K
used_memory_scripts:0
used_memory_scripts_human:0B
number_of_cached_scripts:0
maxmemory:0
maxmemory_human:0B
maxmemory_policy:noeviction
allocator_frag_ratio:1.27
allocator_frag_bytes:234368
allocator_rss_ratio:3.86
allocator_rss_bytes:3158016
rss_overhead_ratio:0.71
rss_overhead_bytes:-1216512
mem_fragmentation_ratio:3.66
mem_fragmentation_bytes:2214264
mem_not_counted_for_evict:0
mem_replication_backlog:0
mem_clients_slaves:0
mem_clients_normal:66616
mem_aof_buffer:0
mem_allocator:jemalloc-5.1.0
active_defrag_running:0
lazyfree_pending_objects:0

# Persistence
loading:0
rdb_changes_since_last_save:6
rdb_bgsave_in_progress:0
rdb_last_save_time:1621328944
rdb_last_bgsave_status:ok
rdb_last_bgsave_time_sec:-1
rdb_current_bgsave_time_sec:-1
rdb_last_cow_size:0
aof_enabled:0
aof_rewrite_in_progress:0
aof_rewrite_scheduled:0
aof_last_rewrite_time_sec:-1
aof_current_rewrite_time_sec:-1
aof_last_bgrewrite_status:ok
aof_last_write_status:ok
aof_last_cow_size:0

# Stats
total_connections_received:47
total_commands_processed:165
instantaneous_ops_per_sec:0
total_net_input_bytes:4157
total_net_output_bytes:39739
instantaneous_input_kbps:0.00
instantaneous_output_kbps:0.00
rejected_connections:0
sync_full:0
sync_partial_ok:0
sync_partial_err:0
expired_keys:0
expired_stale_perc:0.00
expired_time_cap_reached_count:0
evicted_keys:0
keyspace_hits:15
keyspace_misses:0
pubsub_channels:0
pubsub_patterns:0
latest_fork_usec:0
migrate_cached_sockets:0
slave_expires_tracked_keys:0
active_defrag_hits:0
active_defrag_misses:0
active_defrag_key_hits:0
active_defrag_key_misses:0

# Replication
role:master
connected_slaves:0
master_replid:1a81099db358c098605516f9aa8b9fc43a3c6a4a
master_replid2:0000000000000000000000000000000000000000
master_repl_offset:0
second_repl_offset:-1
repl_backlog_active:0
repl_backlog_size:1048576
repl_backlog_first_byte_offset:0
repl_backlog_histlen:0

# CPU
used_cpu_sys:887.840000
used_cpu_user:177.680000
used_cpu_sys_children:0.000000
used_cpu_user_children:0.000000

# Cluster
cluster_enabled:0

# Keyspace
db0:keys=1,expires=0,avg_ttl=0
db1:keys=1,expires=0,avg_ttl=0
`

func TestGetInfo(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mock.Regexp().ExpectInfo(redisInfo)

	input := &Input{
		Service: "dev-test",
		Tags:    make(map[string]string),
		client:  db,
	}

	resData, err := input.collectInfoMeasurement()
	if err != nil {
		t.Log("collect data err", err)
	}

	for _, pt := range resData {
		point, err := pt.LineProto()
		if err != nil {
			fmt.Println("error =======>", err)
		} else {
			fmt.Println("point line =====>", point.String())
		}
	}

}

func TestCollectInfoMeasurement(t *testing.T) {
	input := &Input{
		Host:     "127.0.0.1",
		Port:     6379,
		Password: "dev",
		Service:  "dev-test",
		Tags:     make(map[string]string),
		Log:      &inputs.TailerOption{},
	}

	err := input.initCfg()
	if err != nil {
		t.Log("init cfg err", err)
		return
	}

	resData, err := input.collectInfoMeasurement()
	if err != nil {
		t.Log("collect data err", err)
	}

	for _, pt := range resData {
		point, err := pt.LineProto()
		if err != nil {
			fmt.Println("error =======>", err)
		} else {
			fmt.Println("point line =====>", point.String())
		}
	}
}

func TestCollectClientMeasurement(t *testing.T) {
	input := &Input{
		Host:     "127.0.0.1",
		Port:     6379,
		Password: "dev",
		Service:  "dev-test",
		Tags:     make(map[string]string),
	}

	input.initCfg()

	resData, err := input.collectClientMeasurement()
	if err != nil {
		assert.Error(t, err, "collect data err")
	}

	for _, pt := range resData {
		point, err := pt.LineProto()
		if err != nil {
			fmt.Println("error =======>", err)
		} else {
			fmt.Println("point line =====>", point.String())
		}
	}
}

func TestCollectCommandMeasurement(t *testing.T) {
	input := &Input{
		Host:     "127.0.0.1",
		Port:     6379,
		Password: "dev",
		Service:  "dev-test",
		Tags:     make(map[string]string),
	}

	input.initCfg()

	resData, err := input.collectCommandMeasurement()
	if err != nil {
		assert.Error(t, err, "collect data err")
	}

	for _, pt := range resData {
		point, err := pt.LineProto()
		if err != nil {
			fmt.Println("error =======>", err)
		} else {
			fmt.Println("point line =====>", point.String())
		}
	}
}

func TestCollectSlowlogMeasurement(t *testing.T) {
	input := &Input{
		Host:     "127.0.0.1",
		Port:     6379,
		Password: "dev",
		Service:  "dev-test",
		Tags:     make(map[string]string),
	}

	err := input.initCfg()
	if err != nil {
		t.Log("init cfg err", err)
		return
	}

	resData, err := input.collectSlowlogMeasurement()
	if err != nil {
		assert.Error(t, err, "collect data err")
	}

	for _, pt := range resData {
		point, err := pt.LineProto()
		if err != nil {
			fmt.Println("error =======>", err)
		} else {
			fmt.Println("point line =====>", point.String())
		}
	}
}

func TestCollectBigKeyMeasurement(t *testing.T) {
	input := &Input{
		Host:     "127.0.0.1",
		Port:     6379,
		Password: "dev",
		Service:  "dev-test",
		Tags:     make(map[string]string),
		Keys:     []string{"queue"},
		DB:       1,
	}

	input.initCfg()

	resData, err := input.collectBigKeyMeasurement()
	if err != nil {
		assert.Error(t, err, "collect data err")
	}

	for _, pt := range resData {
		point, err := pt.LineProto()
		if err != nil {
			fmt.Println("error =======>", err)
		} else {
			fmt.Println("point line =====>", point.String())
		}
	}
}

func TestCollectDBMeasurement(t *testing.T) {
	input := &Input{
		Host:     "127.0.0.1",
		Port:     6379,
		Password: "dev",
		Service:  "dev-test",
		Tags:     make(map[string]string),
		Keys:     []string{"queue"},
		DB:       1,
	}

	input.initCfg()

	resData, err := input.collectDBMeasurement()
	if err != nil {
		assert.Error(t, err, "collect data err")
	}

	for _, pt := range resData {
		point, err := pt.LineProto()
		if err != nil {
			fmt.Println("error =======>", err)
		} else {
			fmt.Println("point line =====>", point.String())
		}
	}
}

func TestCollectReplicaMeasurement(t *testing.T) {
	input := &Input{
		Host:     "127.0.0.1",
		Port:     6379,
		Password: "dev",
		Service:  "dev-test",
		Tags:     make(map[string]string),
		Keys:     []string{"queue"},
		DB:       1,
	}

	input.initCfg()

	resData, err := input.collectReplicaMeasurement()
	if err != nil {
		assert.Error(t, err, "collect data err")
	}

	for _, pt := range resData {
		point, err := pt.LineProto()
		if err != nil {
			fmt.Println("error =======>", err)
		} else {
			fmt.Println("point line =====>", point.String())
		}
	}
}

func TestCollect(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		input := &Input{
			Host:     "127.0.0.1",
			Port:     6379,
			Password: "dev",
			Service:  "dev-test",
			Tags:     make(map[string]string),
			Keys:     []string{"queue"},
			DB:       1,
		}

		input.initCfg()
		input.Collect()
	})

	t.Run("error", func(t *testing.T) {
		input := &Input{
			Host:     "127.0.0.1",
			Port:     6379,
			Password: "test",
			Service:  "dev-test",
			Tags:     make(map[string]string),
			Keys:     []string{"queue"},
			DB:       1,
		}

		input.initCfg()
		input.Collect()
	})
}

// func TestRun(t *testing.T) {
// 	input := &Input{
// 		Host:     "127.0.0.1",
// 		Port:     6379,
// 		Password: "dev",
// 		Service:  "dev-test",
// 		Tags:     make(map[string]string),
// 		Keys:     []string{"queue"},
// 		DB:       1,
// 		Log:      &inputs.TailerOption{},
// 	}

// 	input.Run()
// }

// func TestLoadCfg(t *testing.T) {
// 	arr, err := config.LoadInputConfigFile("./redis.conf", func() inputs.Input {
// 		return &Input{}
// 	})

// 	if err != nil {
// 		t.Fatalf("%s", err)
// 	}

// 	arr[0].(*Input).Run()
// }
