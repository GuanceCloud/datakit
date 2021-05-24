### 简介
`Memcached`指标采集，参考`telegraf`的采集方式，基于`tcp`，无侵入性。

### 配置
```
[[inputs.memcached]]
	## 服务器地址，可支持多个
	servers = ["localhost:11211"]
	# unix_sockets = ["/var/run/memcached.sock"]

	## 采集间隔
	# 单位 "ns", "us" (or "µs"), "ms", "s", "m", "h"
	interval = "10s"

	[inputs.memcached.tags]
	# a = "b"
```

### 指标集
当前指标集名称为`memcached`，具体的指标是通过`stats`命令获取的，参考如下。
- accepting_conns
- auth_cmds
- auth_errors
- bytes
- bytes_read
- bytes_written
- cas_badvalmatch
- cas_hits
- cas_misses
- cmd_flush
- cmd_get
- cmd_set
- cmd_touch
- conn_yieldslimit
- connection_structures
- curr_connections
- curr_items
- decr_hits
- decr_misses
- delete_hits
- delete_misses
- evicted_unfetched
- evictions
- expired_unfetchedbefore
- get_hits
- get_misses
- hash_bytes
- hash_is_expanding
- hash_power_level
- incr_hits
- incr_misses
- limit_maxbytes
- listen_disabled_num
- reclaimed
- threads
- total_connections
- total_items
- touch_hits
- touch_misses
- uptime