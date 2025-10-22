// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"strings"
	T "testing"

	"github.com/stretchr/testify/assert"
)

func Test_getConfigAll(t *T.T) {
	testdata := map[string]string{
		"acl-pubsub-default":                  "resetchannels",
		"aclfile":                             "",
		"acllog-max-len":                      "128",
		"active-defrag-cycle-max":             "25",
		"active-defrag-cycle-min":             "1",
		"active-defrag-ignore-bytes":          "104857600",
		"active-defrag-max-scan-fields":       "1000",
		"active-defrag-threshold-lower":       "10",
		"active-defrag-threshold-upper":       "100",
		"active-expire-effort":                "1",
		"activedefrag":                        "no",
		"activerehashing":                     "yes",
		"always-show-logo":                    "no",
		"aof-load-truncated":                  "yes",
		"aof-rewrite-cpulist":                 "",
		"aof-rewrite-incremental-fsync":       "yes",
		"aof-timestamp-enabled":               "no",
		"aof-use-rdb-preamble":                "yes",
		"aof_rewrite_cpulist":                 "",
		"appenddirname":                       "appendonlydir",
		"appendfilename":                      "appendonly.aof",
		"appendfsync":                         "everysec",
		"appendonly":                          "yes",
		"auto-aof-rewrite-min-size":           "67108864",
		"auto-aof-rewrite-percentage":         "100",
		"bgsave-cpulist":                      "",
		"bgsave_cpulist":                      "",
		"bind":                                "198.19.249.182",
		"bind-source-addr":                    "",
		"bio-cpulist":                         "",
		"bio_cpulist":                         "",
		"busy-reply-threshold":                "5000",
		"client-output-buffer-limit":          "normal 0 0 0 slave 268435456 67108864 60 pubsub 33554432 8388608 60",
		"client-query-buffer-limit":           "1073741824",
		"cluster-allow-pubsubshard-when-down": "yes",
		"cluster-allow-reads-when-down":       "no",
		"cluster-allow-replica-migration":     "yes",
		"cluster-announce-bus-port":           "0",
		"cluster-announce-hostname":           "",
		"cluster-announce-human-nodename":     "",
		"cluster-announce-ip":                 "",
		"cluster-announce-port":               "0",
		"cluster-announce-tls-port":           "0",
		"cluster-config-file":                 "nodes.conf",
		"cluster-enabled":                     "yes",
		"cluster-link-sendbuf-limit":          "0",
		"cluster-migration-barrier":           "1",
		"cluster-node-timeout":                "5000",
		"cluster-port":                        "17001",
		"cluster-preferred-endpoint-type":     "ip",
		"cluster-replica-no-failover":         "no",
		"cluster-replica-validity-factor":     "10",
		"cluster-require-full-coverage":       "yes",
		"cluster-slave-no-failover":           "no",
		"cluster-slave-validity-factor":       "10",
		"crash-log-enabled":                   "yes",
		"crash-memcheck-enabled":              "yes",
		"daemonize":                           "no",
		"databases":                           "1",
		"dbfilename":                          "dump.rdb",
		"dir":                                 "/data",
		"disable-thp":                         "yes",
		"dynamic-hz":                          "yes",
		"enable-debug-command":                "no",
		"enable-module-command":               "no",
		"enable-protected-configs":            "no",
		"hash-max-listpack-entries":           "512",
		"hash-max-listpack-value":             "64",
		"hash-max-ziplist-entries":            "512",
		"hash-max-ziplist-value":              "64",
		"hide-user-data-from-log":             "no",
		"hll-sparse-max-bytes":                "3000",
		"hz":                                  "10",
		"ignore-warnings":                     "",
		"io-threads":                          "1",
		"io-threads-do-reads":                 "no",
		"jemalloc-bg-thread":                  "yes",
		"latency-monitor-threshold":           "0",
		"latency-tracking":                    "yes",
		"latency-tracking-info-percentiles":   "50 99 99.9",
		"lazyfree-lazy-eviction":              "no",
		"lazyfree-lazy-expire":                "no",
		"lazyfree-lazy-server-del":            "no",
		"lazyfree-lazy-user-del":              "no",
		"lazyfree-lazy-user-flush":            "no",
		"lfu-decay-time":                      "1",
		"lfu-log-factor":                      "10",
		"list-compress-depth":                 "0",
		"list-max-listpack-size":              "-2",
		"list-max-ziplist-size":               "-2",
		"locale-collate":                      "",
		"logfile":                             "",
		"loglevel":                            "notice",
		"lua-time-limit":                      "5000",
		"masterauth":                          "abc123456",
		"masteruser":                          "",
		"max-new-connections-per-cycle":       "10",
		"max-new-tls-connections-per-cycle":   "1",
		"maxclients":                          "10000",
		"maxmemory":                           "524288000",
		"maxmemory-clients":                   "0",
		"maxmemory-eviction-tenacity":         "10",
		"maxmemory-policy":                    "allkeys-lfu",
		"maxmemory-samples":                   "5",
		"min-replicas-max-lag":                "10",
		"min-replicas-to-write":               "0",
		"min-slaves-max-lag":                  "10",
		"min-slaves-to-write":                 "0",
		"no-appendfsync-on-rewrite":           "no",
		"notify-keyspace-events":              "",
		"oom-score-adj":                       "no",
		"oom-score-adj-values":                "0 200 800",
		"pidfile":                             "",
		"port":                                "7001",
		"proc-title-template":                 "{title} {listen-addr} {server-mode}",
		"propagation-error-behavior":          "ignore",
		"protected-mode":                      "no",
		"proto-max-bulk-len":                  "536870912",
		"rdb-del-sync-files":                  "no",
		"rdb-save-incremental-fsync":          "yes",
		"rdbchecksum":                         "yes",
		"rdbcompression":                      "yes",
		"repl-backlog-size":                   "1048576",
		"repl-backlog-ttl":                    "3600",
		"repl-disable-tcp-nodelay":            "no",
		"repl-diskless-load":                  "disabled",
		"repl-diskless-sync":                  "yes",
		"repl-diskless-sync-delay":            "5",
		"repl-diskless-sync-max-replicas":     "0",
		"repl-ping-replica-period":            "10",
		"repl-ping-slave-period":              "10",
		"repl-timeout":                        "60",
		"replica-announce-ip":                 "",
		"replica-announce-port":               "0",
		"replica-announced":                   "yes",
		"replica-ignore-disk-write-errors":    "no",
		"replica-ignore-maxmemory":            "yes",
		"replica-lazy-flush":                  "no",
		"replica-priority":                    "100",
		"replica-read-only":                   "yes",
		"replica-serve-stale-data":            "yes",
		"replicaof":                           "",
		"requirepass":                         "abc123456",
		"sanitize-dump-payload":               "no",
		"save":                                "3600 1 300 100 60 10000",
		"server-cpulist":                      "",
		"server_cpulist":                      "",
		"set-max-intset-entries":              "512",
		"set-max-listpack-entries":            "128",
		"set-max-listpack-value":              "64",
		"set-proc-title":                      "yes",
		"shutdown-on-sigint":                  "default",
		"shutdown-on-sigterm":                 "default",
		"shutdown-timeout":                    "10",
		"slave-announce-ip":                   "",
		"slave-announce-port":                 "0",
		"slave-ignore-maxmemory":              "yes",
		"slave-lazy-flush":                    "no",
		"slave-priority":                      "100",
		"slave-read-only":                     "yes",
		"slave-serve-stale-data":              "yes",
		"slaveof":                             "",
		"slowlog-log-slower-than":             "5000",
		"slowlog-max-len":                     "1024",
		"socket-mark-id":                      "0",
		"stop-writes-on-bgsave-error":         "yes",
		"stream-node-max-bytes":               "4096",
		"stream-node-max-entries":             "100",
		"supervised":                          "no",
		"syslog-enabled":                      "no",
		"syslog-facility":                     "local0",
		"syslog-ident":                        "redis",
		"tcp-backlog":                         "511",
		"tcp-keepalive":                       "300",
		"timeout":                             "0",
		"tls-auth-clients":                    "yes",
		"tls-ca-cert-dir":                     "",
		"tls-ca-cert-file":                    "",
		"tls-cert-file":                       "",
		"tls-ciphers":                         "",
		"tls-ciphersuites":                    "",
		"tls-client-cert-file":                "",
		"tls-client-key-file":                 "",
		"tls-client-key-file-pass":            "",
		"tls-cluster":                         "no",
		"tls-dh-params-file":                  "",
		"tls-key-file":                        "",
		"tls-key-file-pass":                   "",
		"tls-port":                            "0",
		"tls-prefer-server-ciphers":           "no",
		"tls-protocols":                       "",
		"tls-replication":                     "no",
		"tls-session-cache-size":              "20480",
		"tls-session-cache-timeout":           "300",
		"tls-session-caching":                 "yes",
		"tracking-table-max-keys":             "1000000",
		"unixsocket":                          "",
		"unixsocketperm":                      "0",
		"zset-max-listpack-entries":           "128",
		"zset-max-listpack-value":             "64",
		"zset-max-ziplist-entries":            "128",
		"zset-max-ziplist-value":              "64",
	}

	t.Run("basic", func(t *T.T) {
		ipt := defaultInput()
		addr := "redis.local:7001"

		inst := newInstance()
		inst.ipt = ipt
		inst.addr = addr
		inst.host = "redis.local"

		inst.setup()

		pts := inst.parseConfigAll(testdata)

		for _, pt := range pts {
			assert.Equal(t, "redis.local", pt.Get("host"))
			assert.Equal(t, "redis.local:7001", pt.Get("server"))

			t.Logf("%s", pt.Pretty())

			for k := range testdata {
				if k == "client-output-buffer-limit" {
					continue // this line in seprated point
				}

				if pt.Get("client_output_buffer_limit_bytes") != nil ||
					pt.Get("client_output_buffer_limit_overcome_seconds") != nil {
				} else {
					_k := strings.ReplaceAll(k, "-", "_")
					v := pt.Get(_k)
					assert.NotNilf(t, v, "%s: %v", _k, v)

					switch k {
					case "requirepass", "masterauth", "tls-key-file-pass", "tls-client-key-file-pass":
						assert.Truef(t, pt.Get(_k) == "not-set" || pt.Get(_k) == CREDENTIALSTR, "got %s", pt.Get(_k))
					}
				}
			}
		}
	})
}
