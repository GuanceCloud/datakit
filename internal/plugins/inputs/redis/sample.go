// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

const (
	configSample = `
[[inputs.redis]]
  host = "localhost"
  port = 6379

  ## TLS connection config, redis-cli version must up 6.0+
  ## See also: https://redis.io/docs/latest/operate/oss_and_stack/management/security/encryption/
  # ca_certs             = ["/opt/tls/ca.crt"]
  # cert                 = "/opt/tls/redis.crt"
  # cert_key             = "/opt/tls/redis.key"
  # server_name          = "your-SNI-name"
  # insecure_skip_verify = true
  ## we can encode these file content in base64 format:
  # ca_certs_base64 = ['''LONG_STING......''']
  # cert_base64     = '''LONG_STING......'''
  # cert_key_base64 = '''LONG_STING......'''


  # Use UDS to collect to redis server
  # unix_socket_path = "/var/run/redis/redis.sock"

  # Redis user/password
  # username = "<USERNAME>"
  # password = "<PASSWORD>"

  connect_timeout = "10s"

  ## Configure multiple dbs.
  dbs = [0]

  ## metric collect interval.
  interval = "15s"

  ## topology refresh interval for cluster/sentinel mode (default 10m)
  topology_refresh_interval = "10m"

  ## v2+ override all measurement names to "redis"
  measurement_version = "v2"

  # Slog log configurations.
  slow_log = true       # default enable collect slow log
  all_slow_log = false  # Collect all slowlogs returned by Redis
  slowlog_max_len = 128 # only collect the top-128 slow logs

  # Collect INFO LATENCYSTATS output as metrics with quantiles.
  latency_percentiles = false

  # Default enable election on redis collection.
  election = true

  # For cluster redis
  # [inputs.redis.cluster]
  #   hosts = [ "localhost:6379" ]

  # For master-slave redis
  # [inputs.redis.master_slave]
  #   hosts       = [ "localhost:26380" ] # master or/and slave ip/host
  #   [inputs.redis.master_slave.sentinel]
  #     hosts       = [ "localhost:26380" ] # sentinel ip/host
  #     master_name = "your-master-name"
  #     password    = "sentinel-password"

  # Collect hot and big keys
  [inputs.redis.hot_big_keys]
    enable                      = false
    top_n                       = 10        # report top N big and hot keys, default 10
    big_key_interval            = "3h"      # scan big keys(length and mem usage) interval, default 3 hours
    hot_key_interval            = "15m"     # scan hot keys interval, default 15 minutes
    scan_sleep                  = "200ms"   # sleep every 100 batches to reduce CPU impact on Redis server
    scan_batch_size             = 100       # scan 100 keys on each iteration
    bigkey_threshold_bytes      = 10485760  # keys larger than 10MiB
    bigkey_threshold_len        = 5000      # or elements larger than 5000 are considered to be BIG keys.
    mem_usage_samples           = 100       # collect key's memory usage by sample 100(MEMORY USAGE <key> SAMPLES 100)
    target_role = "master"                  # target role for scanning: "master" or "replica". standalone and cluster modes only support "master"

  # Collect redis client list
  [inputs.redis.collect_client_list]
    log_on_flags = "bxOR" # For more flag info, see: https://redis.io/docs/latest/commands/client-list/
  
  # Config on collecting Redis logging on disk
  #[inputs.redis.log]
  # files              = ["/var/log/redis/*.log"] # required, glob logfiles
  # ignore             = [""]
  # pipeline           = "redis.p"                # grok pipeline script path
  # character_encoding = ""                       # default empty, optionals: "utf-8"/"utf-16le"/"utf-16le"/"gbk"/"gb18030"
  # match              = '''^\S.*'''

  [inputs.redis.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`

	pipelineCfg = `
add_pattern("date2", "%{MONTHDAY} %{MONTH} %{YEAR}?%{TIME}")

grok(_, "%{INT:pid}:%{WORD:role} %{date2:time} %{NOTSPACE:serverity} %{GREEDYDATA:msg}")

group_in(serverity, ["."], "debug", status)
group_in(serverity, ["-"], "verbose", status)
group_in(serverity, ["*"], "notice", status)
group_in(serverity, ["#"], "warning", status)

cast(pid, "int")
default_time(time)
`
)
