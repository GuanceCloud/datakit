// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

// extraRealScriptSamples covers alternate parser branches and boundary values in
// the generated integration scripts. The common harness also adds unmatched,
// missing-message, pre-existing-key, and invalid-time cases for every script.
func extraRealScriptSamples(script string) []string {
	switch script {
	case "apache.p":
		return []string{
			`127.0.0.1 - - [17/May/2021:14:51:09 +0800] "-" 304 0`,
		}
	case "consul.p":
		return []string{
			`Dec 31 23:59:59 node-1 consul[1]: 2024-12-31T23:59:59.999+0000 [ERROR] agent.server: final boundary event`,
		}
	case "dameng.p":
		return []string{
			`2024-02-29 23:59:59.999 [WARNING] database P0000000001 T0000000000000000001  leap-day boundary`,
		}
	case "elasticsearch.p":
		return []string{
			`[2021-06-01T11:45:15,927][INFO ][o.e.n.Node               ] [master] started`,
			`[2021-06-01T11:56:06,712][WARN ][i.s.s.query              ] [master] [shopping][0] took[1ms], took_millis[9223372036854775807], total_hits[0 hits]`,
		}
	case "jenkins.p":
		return []string{
			"2021-05-18 03:08:58.053+0000 [id=1]\tWARNING\thudson.Main#run: warning branch",
			"2021-05-18 03:08:58.053+0000 [id=2]\tDEBUG\thudson.Main#run: ungrouped status boundary",
		}
	case "kafka.p":
		return []string{
			`2021-01-01 00:00:00 [main] INFO  KafkaServer - first parser branch`,
			`123 [main] WARN KafkaServer - duration-only second parser branch`,
			`2021-01-01 00:00:00 ERROR KafkaServer:42 - third parser branch`,
			`[2021-01-01 00:00:00,000] INFO fourth parser lower boundary (kafka.server.KafkaServer)`,
		}
	case "kingbase.p":
		return []string{
			`2024-02-29 23:59:59.999 CST [1] WARNING:  leap-day CST boundary`,
		}
	case "mongodb.p":
		return []string{
			`{"t":{"$date":"1970-01-01T00:00:00.000+00:00"},"s":"E","c":"NETWORK","ctx":"main","msg":"epoch boundary"}`,
			`{"t":{},"s":"I","c":"STORAGE","ctx":"main","msg":"missing nested date"}`,
			`{"unterminated":`,
		}
	case "mysql.p":
		return []string{
			`2020-08-06T14:25:02.936146Z 0 [Warning] second parser branch`,
			`20200806 14:25:02 [ERROR] third parser branch`,
			`1970-01-01T00:00:00.000000Z         0 Query     SELECT 1`,
		}
	case "nginx.p":
		return []string{
			`127.0.0.1 - - [24/Mar/2021:13:54:19 +0800] "GET /redirect HTTP/1.1" 300 0 "-" "curl/8.0"`,
			`127.0.0.1 - - [24/Mar/2021:13:54:19 +0800] "GET /missing HTTP/1.1" 400 1 "-" "curl/8.0"`,
			`127.0.0.1 - - [24/Mar/2021:13:54:19 +0800] "GET /failed HTTP/1.1" 500 2 "-" "curl/8.0"`,
			`127.0.0.1 - - [24/Mar/2021:13:54:19 +0800] "GET /outside HTTP/1.1" 600 3 "-" "curl/8.0"`,
		}
	case "postgresql.p":
		return []string{
			`2024-02-29 23:59:59.999 CST [1] LOG: leap-day CST boundary`,
			`2024-02-29 23:59:59.999 UTC [1] WARNING: metadata-free default branch`,
		}
	case "rabbitmq.p":
		return []string{
			`1970-01-01 00:00:00.000 [error] <0.1.0> epoch boundary`,
		}
	case "redis.p":
		return []string{
			`1:M 01 Jan 1970 00:00:00.000 . debug severity`,
			`2:S 29 Feb 2024 23:59:59.999 - verbose severity`,
			`3:M 29 Feb 2024 23:59:59.999 * notice severity`,
			`4:S 29 Feb 2024 23:59:59.999 # warning severity`,
		}
	case "solr.p":
		return []string{
			`1970-01-01 00:00:00.000 INFO  (qtp1-1) [   ] o.a.s.s.HttpSolrCall [core] webapp=null path=null params={} hits=0 status=0 QTime=0`,
		}
	case "sqlserver.p":
		return []string{
			`1970-01-01 00:00:00.00 Server      epoch boundary`,
		}
	case "tdengine.p":
		return []string{
			`08/22 13:44:34.290731 01081508 TAOS_ADAPTER warn "| 300 |    0ns |     127.0.0.1 | GET | /health " model=web`,
			`08/22 13:44:34.290731 01081508 TAOS_ADAPTER fatal "" error_code=1 error_msg=fatal boundary`,
		}
	case "tomcat.p":
		return []string{
			`127.0.0.1 - - [24/Feb/2015:15:57:10 +0530] "GET /redirect HTTP/1.1" 300 0`,
			`127.0.0.1 - - [24/Feb/2015:15:57:10 +0530] "GET /missing HTTP/1.1" 400 1`,
			`127.0.0.1 - - [24/Feb/2015:15:57:10 +0530] "GET /failed HTTP/1.1" 500 2`,
			`127.0.0.1 - - [24/Feb/2015:15:57:10 +0530] "GET /outside HTTP/1.1" 600 3`,
		}
	default:
		return nil
	}
}
