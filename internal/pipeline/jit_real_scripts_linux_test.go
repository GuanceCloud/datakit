// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/lang/platypus"
	"github.com/GuanceCloud/pipeline-go/ptinput"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

type realScriptCorpus struct {
	script  string
	samples []string
}

func TestJITRealIntegrationScriptsMatchPipelineGo(t *testing.T) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	scriptDirectory := os.Getenv("PIPELINE_SCRIPT_DIRECTORY")
	if runtimePath == "" {
		t.Fatal("requires PLATYPUS_JIT_RUNTIME; skipping is not native compatibility evidence")
	}
	if scriptDirectory == "" {
		scriptDirectory = filepath.Join("testdata", "jit-real-scripts")
	}
	runner, err := pljit.NewRunnerWithHost(runtimePath, "pipeline-go-1.4.3-datakit", 64, pljit.NewPipelineGoHost(nil))
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	corpora := realIntegrationCorpora()

	for _, corpus := range corpora {
		t.Run(corpus.script, func(t *testing.T) {
			sourceBytes, err := os.ReadFile(filepath.Join(scriptDirectory, corpus.script))
			if err != nil {
				t.Fatal(err)
			}
			source := string(sourceBytes)
			check := runner.Check(source)
			if strings.Contains(source, "default_time(") {
				expectedStaticMode := "slots"
				if check.Route != pljit.RouteJITNative || check.Capabilities.RequiredHostFlags != 0 ||
					len(check.Capabilities.HostCalls) != 0 ||
					check.Capabilities.InputProjection.Mode != "keys" || check.Capabilities.StaticOutput.Mode != expectedStaticMode {
					t.Fatalf("default_time script is not fully native: %#v", check)
				}
			} else if check.Route != pljit.RouteJITNative || check.Capabilities.RequiredHostFlags != 0 || len(check.Capabilities.HostCalls) != 0 {
				t.Fatalf("script is not purely native: %#v", check)
			}
			if check.Capabilities.Backend == pljit.ExecutionBackendMachineCode &&
				check.Capabilities.ExecutionTier != pljit.ExecutionTierMachineCodeHelper &&
				check.Capabilities.ExecutionTier != pljit.ExecutionTierMachineCodeSlots {
				t.Fatalf("machine-code script has invalid execution tier: %#v", check.Capabilities)
			}
			if corpus.script == "nginx.p" &&
				(check.Capabilities.Backend != pljit.ExecutionBackendMachineCode ||
					check.Capabilities.ExecutionTier != pljit.ExecutionTierMachineCodeHelper) {
				t.Fatalf("nginx must currently use Cranelift helper machine code, got backend=%q tier=%q",
					check.Capabilities.Backend, check.Capabilities.ExecutionTier)
			}
			script, err := NewPlScriptSimple(point.Logging, corpus.script, source)
			if err != nil {
				t.Fatal(err)
			}
			projection, err := runner.Projection(source)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(source, "default_time(") {
				projected := make(map[string]bool, len(projection.Keys()))
				for _, key := range projection.Keys() {
					projected[key] = true
				}
				for _, key := range []string{"status", "time", "pl_msg"} {
					if !projected[key] {
						t.Fatalf("default_time script projection keys = %v, missing %q", projection.Keys(), key)
					}
				}
			}
			samples := append([]string{}, corpus.samples...)
			samples = append(samples, extraRealScriptSamples(corpus.script)...)
			samples = append(samples, "not a recognized log format")
			t.Run("mixed-batches", func(t *testing.T) {
				compareRealScriptMixedBatches(t, runner, script, source, projection, corpus.script, samples)
			})
			for index, message := range samples {
				t.Run(testSampleName(index, message), func(t *testing.T) {
					compareRealScriptPoint(t, runner, script, source, projection, corpus.script, map[string]any{
						"message": message, "sentinel": "keep", "status": "unknown",
					})
				})
			}
			t.Run("missing-message", func(t *testing.T) {
				compareRealScriptPoint(t, runner, script, source, projection, corpus.script, map[string]any{
					"sentinel": "keep", "status": "unknown",
				})
			})
			if strings.Contains(source, "default_time(") {
				t.Run("invalid-existing-time-native", func(t *testing.T) {
					compareRealScriptPoint(t, runner, script, source, projection, corpus.script, map[string]any{
						"message": "not a recognized log format", "time": "not-a-time", "sentinel": "keep",
					})
				})
			}
		})
	}
}

func compareRealScriptMixedBatches(t *testing.T, runner *pljit.Runner, script *platypus.PlScript, source string, projection pljit.InputProjection, measurement string, samples []string) {
	t.Helper()
	fixtures := make([]map[string]any, 0, len(samples)+2)
	for _, message := range samples {
		fixtures = append(fixtures, map[string]any{"message": message, "status": "unknown"})
	}
	fixtures = append(fixtures, map[string]any{"status": "unknown"}, map[string]any{"message": "not a recognized log format", "time": "not-a-time"})
	for _, size := range []int{1, 2, 4, 8, 10, 128} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			// Every fixture appears at every batch size; the final batch wraps.
			for offset := 0; offset < len(fixtures); offset += size {
				actual, expected := make([]*point.Point, size), make([]*point.Point, size)
				failed, dropped := make([]bool, size), make([]bool, size)
				for i := range actual {
					fields := make(map[string]any)
					for key, value := range fixtures[(offset+i)%len(fixtures)] {
						fields[key] = value
					}
					fields["sentinel"] = fmt.Sprintf("%d/%d", offset, i)
					actual[i], expected[i] = newRealScriptPoint(measurement, fields), newRealScriptPoint(measurement, fields)
					wrapped := ptinput.PtWrap(point.Logging, expected[i])
					failed[i] = script.Run(wrapped, nil, nil) != nil
					if len(wrapped.GetSubPoint()) != 0 {
						t.Fatal("corpus gained subpoints; extend the mixed oracle before accepting it")
					}
					if !failed[i] {
						dropped[i] = wrapped.Dropped()
						expected[i] = wrapped.Point()
					}
				}
				input, err := encodeProjectedJITPoints(point.Logging, actual, projection)
				if err != nil {
					t.Fatal(err)
				}
				batch, err := runner.Process(source, input)
				if err != nil {
					t.Fatal(err)
				}
				if len(batch.Records) != size {
					t.Fatal("native record count mismatch")
				}
				for i, record := range batch.Records {
					if (record.Status == pljit.TerminalError) != failed[i] || (record.Status != pljit.TerminalOK && record.Status != pljit.TerminalError) {
						t.Fatalf("offset=%d record=%d Go failed=%v native=%v error=%s", offset, i, failed[i], record.Status, record.Error)
					}
					if failed[i] && !record.CommitPrefixError {
						t.Fatalf("Go replay forbidden: %s", record.Error)
					}
					var gotDrop bool
					if batch.Static != nil {
						_, gotDrop, err = applyJITStatic(point.Logging, actual[i], batch.Static, i, nil)
					} else {
						_, gotDrop, err = applyJITRecord(point.Logging, actual[i], record, uint64(i), nil)
					}
					if err != nil {
						t.Fatal(err)
					}
					if !failed[i] {
						ptinput.PtWrap(point.Logging, actual[i]).KeyTime2Time()
					}
					if gotDrop != dropped[i] {
						t.Fatal("drop mismatch")
					}
					if equal, why := actual[i].EqualWithReason(expected[i]); !equal {
						t.Fatalf("offset=%d record=%d: %s", offset, i, why)
					}
				}
			}
		})
	}
}

func compareRealScriptPoint(t *testing.T, runner *pljit.Runner, script *platypus.PlScript, source string, projection pljit.InputProjection, measurement string, fields map[string]any) {
	t.Helper()
	want := newRealScriptPoint(measurement, fields)
	run := pointRun{point: want, script: script, started: time.Now()}
	runPipelineGo(point.Logging, &run, nil)

	got := newRealScriptPoint(measurement, fields)
	input, err := encodeProjectedJITPoints(point.Logging, []*point.Point{got}, projection)
	if err != nil {
		t.Fatal(err)
	}
	batch, err := runner.Process(source, input)
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Records) != 1 {
		t.Fatalf("unexpected JIT terminal: %#v", batch)
	}
	if strings.Contains(source, "default_time(") &&
		batch.Static == nil {
		t.Fatalf("default_time script did not use static output: %#v", batch)
	}
	if batch.Records[0].Status == pljit.TerminalError {
		if batch.Records[0].CommitPrefixError {
			if !strings.Contains(string(batch.Records[0].Error), "E_PIPELINE_SCRIPT_ERROR") {
				t.Fatalf("unexpected committed JIT terminal: %#v", batch)
			}
			if batch.Static != nil {
				_, _, err = applyJITStatic(point.Logging, got, batch.Static, 0, nil)
			} else {
				_, _, err = applyJITRecord(point.Logging, got, batch.Records[0], 0, nil)
			}
			if err != nil {
				t.Fatalf("apply committed JIT prefix: %v", err)
			}
			if run.dropped || run.output == nil {
				t.Fatalf("pipeline-go script error unexpectedly dropped its point: dropped=%v output=%#v", run.dropped, run.output)
			}
			if equal, reason := got.EqualWithReason(run.output); !equal {
				t.Fatalf("committed JIT error differs from pipeline-go: %s\nJIT: %#v\npipeline-go: %#v", reason, got.KVMap(), run.output.KVMap())
			}
			return
		}
		// Go is only the offline oracle above. A compatibility terminal must
		// fail this corpus, never replay the tested Point through Go.
		t.Fatalf("native execution failed; Go replay is forbidden: status=%v error=%s", batch.Records[0].Status, batch.Records[0].Error)
		return
	}
	if batch.Records[0].Status != pljit.TerminalOK {
		t.Fatalf("unexpected JIT terminal: %#v", batch)
	}
	var dropped bool
	if batch.Static != nil {
		_, dropped, err = applyJITStatic(point.Logging, got, batch.Static, 0, nil)
	} else {
		_, dropped, err = applyJITRecord(point.Logging, got, batch.Records[0], 0, nil)
	}
	if err != nil {
		t.Fatal(err)
	}
	ptinput.PtWrap(point.Logging, got).KeyTime2Time()
	if dropped != run.dropped {
		t.Fatalf("dropped=%v, pipeline-go=%v", dropped, run.dropped)
	}
	if run.output == nil {
		if !dropped {
			t.Fatal("pipeline-go returned nil without dropping")
		}
		return
	}
	if equal, reason := got.EqualWithReason(run.output); !equal {
		t.Fatalf("JIT differs from pipeline-go: %s\nJIT: %#v\npipeline-go: %#v", reason, got.KVMap(), run.output.KVMap())
	}
}

func newRealScriptPoint(measurement string, fields map[string]any) *point.Point {
	copyFields := make(map[string]any, len(fields))
	for key, value := range fields {
		copyFields[key] = value
	}
	kvs := point.NewKVs(copyFields).SetTag("existing_tag", "keep")
	return point.NewPoint(measurement, kvs, point.WithTime(time.Unix(1_700_000_000, 123)))
}

func testSampleName(index int, message string) string {
	name := "sample"
	if strings.Contains(message, "not a recognized") {
		name = "unmatched"
	}
	return name + "-" + time.Unix(int64(index), 0).UTC().Format("05")
}

func realIntegrationCorpora() []realScriptCorpus {
	return []realScriptCorpus{
		{script: "apache.p", samples: []string{
			`127.0.0.1 - - [17/May/2021:14:51:09 +0800] "GET /server-status?auto HTTP/1.1" 200 917`,
			`[Tue May 19 18:39:45.272121 2021] [access_compat:error] [pid 9802] [client ::1:50547] AH01797: client denied by server configuration: /Library/WebServer/Documents/server-status`,
			`[Fri Sep 09 10:42:29.902022 2011] [core:error] [pid 35708:tid 4328636416] [client 72.15.99.187] AH00124: Request exceeded the limit of 10 internal redirects due to probable configuration error`,
			`127.0.0.1 - frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`,
		}},
		{script: "consul.p", samples: []string{
			`Sep 18 19:30:23 derrick-ThinkPad-X230 consul[11803]: 2021-09-18T19:30:23.522+0800 [INFO]  agent.server.connect: initialized primary datacenter CA with provider: provider=consul`,
			`2024-12-19T13:29:40.067-0800 [INFO]  agent.server: server starting leave`,
		}},
		{script: "dameng.p", samples: []string{
			`2025-06-23 10:04:27.793 [INFO] dminit P0000008230 T0000000000000008230  INI parameter RECYCLE_POOLS changed, the original value 0, new value 1`,
			`2024-08-01 14:54:43.110 [ERROR] database P0000003743 T0000000000000003743  os_file_open_low_real error! desc: Permission denied, path: /dmdbms/data/itsdata/itsdata01.log, code: 13`,
		}},
		{script: "elasticsearch.p", samples: []string{
			`[2021-06-01T11:45:15,927][WARN ][o.e.c.r.a.DiskThresholdMonitor] [master] high disk watermark [90%] exceeded on [A2kEFgMLQ1-vhMdZMJV3Iw][master][/tmp/elasticsearch-cluster/nodes/0] free: 17.1gb[7.3%], shards will be relocated away from this node`,
			`[2021-06-01T11:56:06,712][WARN ][i.s.s.query              ] [master] [shopping][0] took[36.3ms], took_millis[36], total_hits[5 hits], types[], stats[], search_type[QUERY_THEN_FETCH], total_shards[1]`,
			`[2021-06-01T11:56:19,084][WARN ][i.i.s.index              ] [master] [shopping/X17jbNZ4SoS65zKTU9ZAJg] took[34.1ms], took_millis[34], type[_doc], id[LgC3xXkBLT9WrDT1Dovp]`,
		}},
		{script: "jenkins.p", samples: []string{
			"2021-05-18 03:08:58.053+0000 [id=32]\tINFO\tjenkins.InitReactorRunner$1#onAttained: Started all plugins",
			"2025-12-24 18:57:18.991+0000 [id=33]\tSEVERE\thudson.util.BootFailure#publish: Failed to initialize Jenkins",
		}},
		{script: "kafka.p", samples: []string{
			`[2020-07-07 15:04:29,333] DEBUG Progress event: HTTP_REQUEST_COMPLETED_EVENT, bytes: 0 (io.confluent.connect.s3.storage.S3OutputStream:286)`,
			`2021-11-05 05:56:09,169 INFO [Controller id=0] Processing automatic preferred replica leader election (kafka.controller.KafkaController) [controller-event-thread]`,
		}},
		{script: "kingbase.p", samples: []string{
			`2025-06-17 13:07:10.952 UTC [999] ERROR:  relation "sys_stat_activity" does not exist at character 240`,
			`FATAL: could not receive data from WAL stream: ERROR: requested WAL segment 00000001000003AB00000026 has already been removed`,
		}},
		{script: "mongodb.p", samples: []string{
			`{"t":{"$date":"2021-06-03T09:12:19.977+00:00"},"s":"I","c":"STORAGE","id":22430,"ctx":"WTCheckpointThread","msg":"WiredTiger message"}`,
			`{"t":{"$date":"2020-05-18T20:18:12.734+00:00"},"s":"W","c":"ASIO","id":22601,"ctx":"main","svc":"R","msg":"No TransportLayer configured during NetworkInterface startup"}`,
			`2020-03-05T20:10:30.290+0600 I NETWORK [listener] Listening on /tmp/mongodb-27017.sock`,
		}},
		{script: "mysql.p", samples: []string{
			`2017-12-29T12:33:33.095243Z         2 Query     SELECT TABLE_SCHEMA, TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE CREATE_OPTIONS LIKE '%partitioned%';`,
			"# Time: 2019-11-27T10:43:13.460744Z\n# User@Host: root[root] @ localhost [1.2.3.4]  Id:    35\n# Query_time: 0.214922  Lock_time: 0.000184 Rows_sent: 248832  Rows_examined: 72\n# Thread_id: 55   Killed: 0  Errno: 0\n# Bytes_sent: 123456   Bytes_received: 0\nSET timestamp=1574851393;\nSELECT * FROM fruit f1, fruit f2, fruit f3, fruit f4, fruit f5",
			`2020-08-06T14:25:02.936146Z 0 [Warning] [MY-010068] [Server] CA certificate /var/mysql/sslinfo/cacert.pem is self signed.`,
			`{"prio":1,"err_code":100,"msg":"JSON error sink is not covered by the current script"}`,
		}},
		{script: "nginx.p", samples: []string{
			`127.0.0.1 - - [24/Mar/2021:13:54:19 +0800] "GET /basic_status HTTP/1.1" 200 97 "-" "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.72 Safari/537.36"`,
			`2021/04/21 09:24:04 [alert] 7#7: *168 write() to "/var/log/nginx/access.log" failed (28: No space left on device) while logging request, client: 120.204.196.129, server: localhost, request: "GET / HTTP/1.1", host: "47.98.103.73"`,
			`2021/04/29 16:24:38 [emerg] 50102#0: unexpected ";" in /usr/local/etc/nginx/nginx.conf:23`,
			`2013/07/30 13:40:08 [error] 3343#0: *247 upstream prematurely closed connection while reading response header from upstream, client: 192.241.195.20, server: store.vinioliveira.com, request: "GET /admin HTTP/1.1", upstream: "http://unix:/tmp/store.unicorn.sock:/admin", host: "store.vinioliveira.com"`,
		}},
		{script: "postgresql.p", samples: []string{
			`2021-05-31 15:23:45.110 CST [74305] test [pgAdmin 4 - DB:postgres] postgres [127.0.0.1] 60b48f01.12241 LOG: statement: SELECT 1`,
			`2023-09-13 20:40:27.766 UTC [295538] FATAL: the database system is in recovery mode`,
			`2024-05-15 23:12:01.616 UTC [1299981][client backend][2/2:0] LOG: statement:`,
		}},
		{script: "rabbitmq.p", samples: []string{
			`2021-05-26 14:20:06.105 [warning] <0.12897.46> rabbitmqctl node_health_check and its HTTP API counterpart are DEPRECATED.`,
			"=INFO REPORT==== 11-Oct-2012::12:53:54 ===\naccepting AMQP connection <0.304.0> (127.0.0.1:46381 -> 127.0.0.1:5672)",
		}},
		{script: "redis.p", samples: []string{
			`122:M 14 May 2019 19:11:40.164 * Background saving terminated with success`,
			`1516:S 24 Jan 16:04:57.905 * Background saving terminated with success`,
			`2817:M 23 Nov 2024 11:29:00.549 # Background saving terminated by signal 31`,
			`2262:signal-handler (1614036707) Received SIGTERM scheduling shutdown...`,
		}},
		{script: "solr.p", samples: []string{
			`2020-02-26 09:10:57.641 INFO  (qtp1635378213-18) [   ] o.a.s.s.HttpSolrCall [admin] webapp=null path=/admin/info/system params={} status=0 QTime=52`,
			`2013-10-01 12:33:08.319 INFO (org.apache.solr.core.SolrCore) [collection1] webapp.reporter`,
		}},
		{script: "sqlserver.p", samples: []string{
			`2020-01-01 00:00:01.00 spid28s     Server is listening on [ ::1 <ipv6> 1431] accept sockets 1.`,
			`2020-01-01 00:00:02.00 Server      Common language runtime (CLR) functionality initialized.`,
			`2023-03-10 09:38:46.73 spid9s SQL Server is terminating in response to a 'stop' request from Service Control Manager.`,
		}},
		{script: "tdengine.p", samples: []string{
			`08/22 13:44:34.290731 01081508 TAOS_ADAPTER info "| 204 |    1.641678ms |     172.16.5.29 | POST | /influxdb/v1/write?db=biz " model=web sessionID=48847`,
			`08/22 13:44:34.290731 01081508 TAOS_ADAPTER error "" error_code=512 error_msg=db is not specified`,
			`06/22 08:17:51.155182 00000034 taos_ADAPTER info "| 200 |  10.29068432s |       127.0.0.1 | POST | %2Frest%2Fsql%2Flog " model=web req_id=0x1bd203f05c401b2d`,
			`06/22 08:17:51.155097 00000034 taos_ADAPTER error "" error_code=512 error_msg=db is not specified`,
		}},
		{script: "tomcat.p", samples: []string{
			`0:0:0:0:0:0:0:1 - admin [24/Feb/2015:15:57:10 +0530] "GET /manager/images/tomcat.gif HTTP/1.1" 200 2066`,
			`06-Sep-2021 22:33:30.513 INFO [main] org.apache.catalina.startup.VersionLoggerListener.log Command line argument: -Xmx256m`,
		}},
	}

}
