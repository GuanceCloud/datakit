// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/dustin/go-humanize"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/funcs"
)

type versionDesc struct {
	Version  string `json:"version"`
	Date     string `json:"date_utc"`
	Uploader string `json:"uploader"`
	Branch   string `json:"branch"`
	Commit   string `json:"commit"`
	Go       string `json:"go"`
}

func tarFiles(pubPath, buildPath, appName, goos, goarch string) (string, string) {
	gz := fmt.Sprintf("%s-%s-%s-%s.tar.gz",
		appName, goos, goarch, ReleaseVersion)
	gzPath := filepath.Join(pubPath, ReleaseType, gz)

	args := []string{
		`czf`,
		gzPath,
		`-C`,
		// the whole basePath/appName-<goos>-<goarch> dir
		filepath.Join(buildPath, fmt.Sprintf("%s-%s-%s", appName, goos, goarch)), `.`,
	}

	cmd := exec.Command("tar", args...) //nolint:gosec

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	l.Debugf("tar %s...", gzPath)
	if err := cmd.Run(); err != nil {
		l.Fatal(err)
	}
	return gz, gzPath
}

func generateInstallScript() error {
	x := struct {
		InstallBaseURL string
		Version        string
	}{
		InstallBaseURL: DownloadAddr,
		Version:        ReleaseVersion,
	}

	for k, v := range map[string]string{
		"install.sh.template":   "install.sh",
		"install.ps1.template":  "install.ps1",
		"datakit.yaml.template": "datakit.yaml",
	} {
		txt, err := ioutil.ReadFile(filepath.Clean(k))
		if err != nil {
			return err
		}

		t := template.New("")
		t, err = t.Parse(string(txt))
		if err != nil {
			return err
		}

		fd, err := os.OpenFile(filepath.Clean(v),
			os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
		if err != nil {
			return err
		}

		l.Infof("creating install script %s", v)
		if err := t.Execute(fd, x); err != nil {
			return err
		}

		fd.Close() //nolint:errcheck,gosec
	}

	return nil
}

func generateMetaInfo() error {
	return cmds.ExportMetaInfo("measurements-meta.json")
}

func generatePipelineDoc() error {
	encoding := base64.StdEncoding
	protoPrefix, descPrefix := "函数原型：", "函数说明："
	// Write function description & prototype.
	for _, plDoc := range funcs.PipelineFunctionDocs {
		lines := strings.Split(plDoc.Doc, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, protoPrefix) {
				proto := strings.TrimPrefix(line, protoPrefix)
				// Prototype line contains starting and trailing ` only.
				if len(proto) >= 2 && strings.Index(proto, "`") == 0 && strings.Index(proto[1:], "`") == len(proto[1:])-1 {
					proto = proto[1 : len(proto)-1]
				}
				plDoc.Prototype = proto
			} else if strings.HasPrefix(line, descPrefix) {
				plDoc.Description = strings.TrimPrefix(line, descPrefix)
			}
		}
	}
	// Encode Markdown docs with base64.
	for _, plDoc := range funcs.PipelineFunctionDocs {
		plDoc.Doc = encoding.EncodeToString([]byte(plDoc.Doc))
		plDoc.Prototype = encoding.EncodeToString([]byte(plDoc.Prototype))
		plDoc.Description = encoding.EncodeToString([]byte(plDoc.Description))
	}
	exportPLDocs := struct {
		Version   string                  `json:"version"`
		Docs      string                  `json:"docs"`
		Functions map[string]*funcs.PLDoc `json:"functions"`
	}{
		Version:   git.Version,
		Docs:      "经过 base64 编码的 pipeline 函数文档，包括各函数原型、函数说明、使用示例",
		Functions: funcs.PipelineFunctionDocs,
	}
	data, err := json.Marshal(exportPLDocs)
	if err != nil {
		return err
	}
	f, err := os.Create("pipeline-docs.json")
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck,gosec
	if _, err := f.Write(data); err != nil {
		return err
	}
	return nil
}

//nolint:lll
func generatePipelineScripts() error {
	encoding := base64.StdEncoding
	type pipelineDemo struct {
		Pipeline string            `json:"pipeline"`
		Examples map[string]string `json:"examples"`
	}
	logExamples := map[string]map[string]string{
		"apache.p": {
			"Apache error log":  `[Tue May 19 18:39:45.272121 2021] [access_compat:error] [pid 9802] [client ::1:50547] AH01797: client denied by server configuration: /Library/WebServer/Documents/server-status`,
			"Apache access log": `127.0.0.1 - - [17/May/2021:14:51:09 +0800] "GET /server-status?auto HTTP/1.1" 200 917`,
		},

		"consul.p": {
			"Consul log": `Sep 18 19:30:23 derrick-ThinkPad-X230 consul[11803]: 2021-09-18T19:30:23.522+0800 [INFO]  agent.server.connect: initialized primary datacenter CA with provider: provider=consul`,
		},
		"elasticsearch.p": {
			"ElasticSearch log":             `[2021-06-01T11:45:15,927][WARN ][o.e.c.r.a.DiskThresholdMonitor] [master] high disk watermark [90%] exceeded on [A2kEFgMLQ1-vhMdZMJV3Iw][master][/tmp/elasticsearch-cluster/nodes/0] free: 17.1gb[7.3%], shards will be relocated away from this node; currently relocating away shards totalling [0] bytes; the node is expected to continue to exceed the high disk watermark when these relocations are complete`,
			"ElasticSearch search slow log": `[2021-06-01T11:56:06,712][WARN ][i.s.s.query              ] [master] [shopping][0] took[36.3ms], took_millis[36], total_hits[5 hits], types[], stats[], search_type[QUERY_THEN_FETCH], total_shards[1], source[{"query":{"match":{"name":{"query":"Nariko","operator":"OR","prefix_length":0,"max_expansions":50,"fuzzy_transpositions":true,"lenient":false,"zero_terms_query":"NONE","auto_generate_synonyms_phrase_query":true,"boost":1.0}}},"sort":[{"price":{"order":"desc"}}]}], id[],`,
			"ElasticSearch index slow log":  `[2021-06-01T11:56:19,084][WARN ][i.i.s.index              ] [master] [shopping/X17jbNZ4SoS65zKTU9ZAJg] took[34.1ms], took_millis[34], type[_doc], id[LgC3xXkBLT9WrDT1Dovp], routing[], source[{"price":222,"name":"hello"}]`,
		},
		"jenkins.p": {
			"Jenkins log": `2021-05-18 03:08:58.053+0000 [id=32]	INFO	jenkins.InitReactorRunner$1#onAttained: Started all plugins`,
		},
		"kafka.p": {
			"Kafka log": `[2020-07-07 15:04:29,333] DEBUG Progress event: HTTP_REQUEST_COMPLETED_EVENT, bytes: 0 (io.confluent.connect.s3.storage.S3OutputStream:286)`,
		},
		"mongodb.p": {
			"MongoDB log": `{"t":{"$date":"2021-06-03T09:12:19.977+00:00"},"s":"I",  "c":"STORAGE",  "id":22430,   "ctx":"WTCheckpointThread","msg":"WiredTiger message","attr":{"message":"[1622711539:977142][1:0x7f1b9f159700], WT_SESSION.checkpoint: [WT_VERB_CHECKPOINT_PROGRESS] saving checkpoint snapshot min: 653, snapshot max: 653 snapshot count: 0, oldest timestamp: (0, 0) , meta checkpoint timestamp: (0, 0)"}}`,
		},
		"mysql.p": {
			"MySQL log": `2017-12-29T12:33:33.095243Z         2 Query     SELECT TABLE_SCHEMA, TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE CREATE_OPTIONS LIKE '%partitioned%';`,
			"MySQL slow log": `# Time: 2019-11-27T10:43:13.460744Z
# User@Host: root[root] @ localhost [1.2.3.4]  Id:    35
# Query_time: 0.214922  Lock_time: 0.000184 Rows_sent: 248832  Rows_examined: 72
# Thread_id: 55   Killed: 0  Errno: 0
# Bytes_sent: 123456   Bytes_received: 0
SET timestamp=1574851393;
SELECT * FROM fruit f1, fruit f2, fruit f3, fruit f4, fruit f5`,
		},
		"nginx.p": {
			"Nginx error log1": `2021/04/21 09:24:04 [alert] 7#7: *168 write() to "/var/log/nginx/access.log" failed (28: No space left on device) while logging request, client: 120.204.196.129, server: localhost, request: "GET / HTTP/1.1", host: "47.98.103.73"`,
			"Nginx error log2": `2021/04/29 16:24:38 [emerg] 50102#0: unexpected ";" in /usr/local/etc/nginx/nginx.conf:23`,
			"Nginx access log": `127.0.0.1 - - [24/Mar/2021:13:54:19 +0800] "GET /basic_status HTTP/1.1" 200 97 "-" "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.72 Safari/537.36"`,
		},
		"postgresql.p": {
			"PostgreSQL log": `2021-05-31 15:23:45.110 CST [74305] test [pgAdmin 4 - DB:postgres] postgres [127.0.0.1] 60b48f01.12241 LOG: statement: 		SELECT psd.*, 2^31 - age(datfrozenxid) as wraparound, pg_database_size(psd.datname) as pg_database_size 		FROM pg_stat_database psd 		JOIN pg_database pd ON psd.datname = pd.datname 		WHERE psd.datname not ilike 'template%' AND psd.datname not ilike 'rdsadmin' 		AND psd.datname not ilike 'azure_maintenance' AND psd.datname not ilike 'postgres'`,
		},
		"rabbitmq.p": {
			"RabbitMQ log": `2021-05-26 14:20:06.105 [warning] <0.12897.46> rabbitmqctl node_health_check and its HTTP API counterpart are DEPRECATED. See https://www.rabbitmq.com/monitoring.html#health-checks for replacement options.`,
		},
		"redis.p": {
			"Redis log": `122:M 14 May 2019 19:11:40.164 * Background saving terminated with success`,
		},
		"solr.p": {
			"Solr log": `2013-10-01 12:33:08.319 INFO (org.apache.solr.core.SolrCore) [collection1] webapp.reporter`,
		},
		"sqlserver.p": {
			"SQLServer log": `2021-05-28 10:46:07.78 spid10s     0 transactions rolled back in database 'msdb' (4:0). This is an informational message only. No user action is required`,
		},
		"tomcat.p": {
			"Tomcat access log":   `0:0:0:0:0:0:0:1 - admin [24/Feb/2015:15:57:10 +0530] "GET /manager/images/tomcat.gif HTTP/1.1" 200 2066`,
			"Tomcat Catalina log": `06-Sep-2021 22:33:30.513 INFO [main] org.apache.catalina.startup.VersionLoggerListener.log Command line argument: -Xmx256m`,
		},
	}
	// Encode log examples with base64.
	for _, examples := range logExamples {
		for k, v := range examples {
			examples[k] = encoding.EncodeToString([]byte(v))
		}
	}
	scriptMap, err := config.GetScriptMap(false)
	if err != nil {
		return err
	}
	// Encode Markdown docs with base64.
	for name, script := range scriptMap {
		scriptMap[name] = encoding.EncodeToString([]byte(script))
	}
	scriptAndExamples := map[string]pipelineDemo{}
	for name, script := range scriptMap {
		scriptAndExamples[name] = pipelineDemo{
			Pipeline: script,
			Examples: logExamples[name],
		}
	}

	data, err := json.Marshal(scriptAndExamples)
	if err != nil {
		return err
	}
	f, err := os.Create("internal-pipelines.json")
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck,gosec
	if _, err := f.Write(data); err != nil {
		return err
	}
	return nil
}

func addOSSFiles(ossPath string, files map[string]string) map[string]string {
	res := map[string]string{}
	for k, v := range files {
		res[path.Join(ossPath, k)] = v
	}
	return res
}

//nolint:funlen,gocyclo
func PubDatakit() error {
	start := time.Now()
	var ak, sk, bucket, ossHost string

	// 在你本地设置好这些 oss-key 环境变量
	switch ReleaseType {
	case ReleaseTesting, ReleaseProduction, ReleaseLocal:
		tag := strings.ToUpper(ReleaseType)
		ak = os.Getenv(tag + "_OSS_ACCESS_KEY")
		sk = os.Getenv(tag + "_OSS_SECRET_KEY")
		bucket = os.Getenv(tag + "_OSS_BUCKET")
		ossHost = os.Getenv(tag + "_OSS_HOST")
	default:
		return fmt.Errorf("unknown release type: %s", ReleaseType)
	}

	if ak == "" || sk == "" {
		return fmt.Errorf("OSS %s/%s not set",
			strings.ToUpper(ReleaseType)+"_OSS_ACCESS_KEY",
			strings.ToUpper(ReleaseType)+"_OSS_SECRET_KEY")
	}

	ossSlice := strings.SplitN(DownloadAddr, "/", 2) // at least 2 parts
	if len(ossSlice) != 2 {
		return fmt.Errorf("invalid download addr: %s", DownloadAddr)
	}
	OSSPath = ossSlice[1]

	oc := &cliutils.OssCli{
		Host:       ossHost,
		PartSize:   512 * 1024 * 1024,
		AccessKey:  ak,
		SecretKey:  sk,
		BucketName: bucket,
		WorkDir:    OSSPath,
	}

	if err := oc.Init(); err != nil {
		return err
	}

	// upload all build archs
	curArchs = parseArchs(Archs)

	if err := generateInstallScript(); err != nil {
		return err
	}

	if err := generateMetaInfo(); err != nil {
		return err
	}

	if err := generatePipelineDoc(); err != nil {
		return err
	}

	if err := generatePipelineScripts(); err != nil {
		return err
	}

	basics := map[string]string{
		"version":                 path.Join(PubDir, ReleaseType, "version"),
		"datakit.yaml":            "datakit.yaml",
		"install.sh":              "install.sh",
		"install.ps1":             "install.ps1",
		"measurements-meta.json":  "measurements-meta.json",
		"pipeline-docs.json":      "pipeline-docs.json",
		"internal-pipelines.json": "internal-pipelines.json",
		fmt.Sprintf("datakit-%s.yaml", ReleaseVersion): "datakit.yaml",
		fmt.Sprintf("install-%s.sh", ReleaseVersion):   "install.sh",
		fmt.Sprintf("install-%s.ps1", ReleaseVersion):  "install.ps1",
	}

	// tar files and collect OSS upload/backup info
	for _, arch := range curArchs {
		parts := strings.Split(arch, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid arch: %s", arch)
		}
		goos, goarch := parts[0], parts[1]

		gzName, gzPath := tarFiles(PubDir, BuildDir, AppName, parts[0], parts[1])
		// gzName := fmt.Sprintf("%s-%s-%s.tar.gz", AppName, goos+"-"+goarch, ReleaseVersion)
		basics[gzName] = gzPath

		for _, appName := range StandaloneApps {
			buildPath := filepath.Join(BuildDir, "standalone")
			switch appName {
			case "datakit-ebpf":
				if parts[0] != runtime.GOOS {
					continue
				}
				if parts[0] != "linux" {
					continue
				}
				if parts[1] != runtime.GOARCH {
					continue
				}
			default:
			}
			gz, gzP := tarFiles(PubDir, buildPath, appName, parts[0], parts[1])
			basics[gz] = gzP
		}

		installerExe := fmt.Sprintf("installer-%s-%s", goos, goarch)
		installerExeWithVer := fmt.Sprintf("installer-%s-%s-%s", goos, goarch, ReleaseVersion)
		if parts[0] == datakit.OSWindows {
			installerExe = fmt.Sprintf("installer-%s-%s.exe", goos, goarch)
			installerExeWithVer = fmt.Sprintf("installer-%s-%s-%s.exe", goos, goarch, ReleaseVersion)
		}

		basics[gzName] = gzPath
		basics[installerExe] = path.Join(PubDir, ReleaseType, installerExe)
		basics[installerExeWithVer] = path.Join(PubDir, ReleaseType, installerExe)
	}

	// Darwin release not under CI, so disable upload `version' file under darwin,
	// only upload darwin related files.
	if Archs == datakit.OSArchDarwinAmd64 && runtime.GOOS == datakit.OSDarwin {
		delete(basics, "version")
	}

	ossfiles := addOSSFiles(OSSPath, basics)

	// test if all file ok before uploading
	for _, k := range ossfiles {
		if _, err := os.Stat(k); err != nil {
			return err
		}
	}

	for k, v := range ossfiles {
		fi, _ := os.Stat(v)
		l.Debugf("%s => %s(%s)...", v, k, humanize.Bytes(uint64(fi.Size())))

		if err := oc.Upload(v, k); err != nil {
			return err
		}
	}

	l.Infof("Done!(elapsed: %v)", time.Since(start))
	return nil
}

//nolint:funlen,gocyclo
func PubDatakitEBpf() error {
	start := time.Now()
	var ak, sk, bucket, ossHost string

	// 在你本地设置好这些 oss-key 环境变量
	switch ReleaseType {
	case ReleaseTesting, ReleaseProduction, ReleaseLocal:
		tag := strings.ToUpper(ReleaseType)
		ak = os.Getenv(tag + "_OSS_ACCESS_KEY")
		sk = os.Getenv(tag + "_OSS_SECRET_KEY")
		bucket = os.Getenv(tag + "_OSS_BUCKET")
		ossHost = os.Getenv(tag + "_OSS_HOST")
	default:
		return fmt.Errorf("unknown release type: %s", ReleaseType)
	}

	if ak == "" || sk == "" {
		return fmt.Errorf("OSS %s/%s not set",
			strings.ToUpper(ReleaseType)+"_OSS_ACCESS_KEY",
			strings.ToUpper(ReleaseType)+"_OSS_SECRET_KEY")
	}

	ossSlice := strings.SplitN(DownloadAddr, "/", 2) // at least 2 parts
	if len(ossSlice) != 2 {
		return fmt.Errorf("invalid download addr: %s", DownloadAddr)
	}
	OSSPath = ossSlice[1]

	oc := &cliutils.OssCli{
		Host:       ossHost,
		PartSize:   512 * 1024 * 1024,
		AccessKey:  ak,
		SecretKey:  sk,
		BucketName: bucket,
		WorkDir:    OSSPath,
	}

	if err := oc.Init(); err != nil {
		return err
	}

	// upload all build archs
	curTmpArchs := parseArchs(Archs)

	basics := map[string]string{}

	// tar files and collect OSS upload/backup info
	for _, arch := range curTmpArchs {
		parts := strings.Split(arch, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid arch: %s", arch)
		}

		for _, appName := range StandaloneApps {
			buildPath := filepath.Join(BuildDir, "standalone")
			switch appName {
			case "datakit-ebpf":
				if parts[0] != runtime.GOOS {
					continue
				}
				if parts[0] != "linux" {
					continue
				}
				if parts[1] != runtime.GOARCH {
					continue
				}
			default:
			}
			curEBpfArchs = append(curEBpfArchs, arch)
			gz, gzP := tarFiles(PubDir, buildPath, appName, parts[0], parts[1])
			basics[gz] = gzP
		}
	}

	ossfiles := addOSSFiles(OSSPath, basics)

	// test if all file ok before uploading
	for _, k := range ossfiles {
		if _, err := os.Stat(k); err != nil {
			return err
		}
	}

	for k, v := range ossfiles {
		fi, _ := os.Stat(v)
		l.Debugf("%s => %s(%s)...", v, k, humanize.Bytes(uint64(fi.Size())))

		if err := oc.Upload(v, k); err != nil {
			return err
		}
	}

	l.Infof("Done!(elapsed: %v)", time.Since(start))
	return nil
}
