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
  # unix_socket_path = "/var/run/redis/redis.sock"
  # 配置多个 db，配置了 dbs，db 也会放入采集列表。dbs=[] 或者不配置则会采集 Redis 中所有非空的 db
  # dbs=[]
  # username = "<USERNAME>"
  # password = "<PASSWORD>"

  ## @param connect_timeout - number - optional - default: 10s
  # connect_timeout = "10s"

  ## @param service - string - optional
  service = "redis"

  ## @param interval - number - optional - default: 15
  interval = "15s"

  ## @param keys - list of strings - optional
  ## The length is 1 for strings.
  ## The length is zero for keys that have a type other than list, set, hash, or sorted set.
  #
  # keys = ["KEY_1", "KEY_PATTERN"]

  ## @param warn_on_missing_keys - boolean - optional - default: true
  ## If you provide a list of 'keys', set this to true to have the Agent log a warning
  ## when keys are missing.
  #
  # warn_on_missing_keys = true

  ## @param slow_log - boolean - optional - default: true
  slow_log = true

  ## @param all_slow_log - boolean - optional - default: false
  ## Collect all slowlogs returned by Redis. When set to false, will only collect slowlog
  ## that are generated after this input starts, and collect the same slowlog only once.
  all_slow_log = false

  ## @param slowlog-max-len - integer - optional - default: 128
  slowlog-max-len = 128

  ## @param command_stats - boolean - optional - default: false
  ## Collect INFO COMMANDSTATS output as metrics.
  # command_stats = false

  ## @param latency_percentiles - boolean - optional - default: false
  ## Collect INFO LATENCYSTATS output as metrics.
  # latency_percentiles = false

  ## Set true to enable election
  election = true

  # [inputs.redis.log]
  # #required, glob logfiles
  # files = ["/var/log/redis/*.log"]

  ## glob filteer
  #ignore = [""]

  ## grok pipeline script path
  #pipeline = "redis.p"

  ## optional encodings:
  ##    "utf-8", "utf-16le", "utf-16le", "gbk", "gb18030" or ""
  #character_encoding = ""

  ## The pattern should be a regexp. Note the use of '''this regexp'''
  ## regexp link: https://golang.org/pkg/regexp/syntax/#hdr-Syntax
  #match = '''^\S.*'''

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
