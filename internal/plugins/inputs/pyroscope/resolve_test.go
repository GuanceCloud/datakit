// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pyroscope

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/profile/metrics"
)

func TestMetadata(t *testing.T) {
	md := &Metadata{
		Format:        metrics.Collapsed,
		Profiler:      metrics.Pyroscope,
		Attachments:   []string{"main.jfr", "metrics.json"},
		Language:      metrics.Golang,
		TagsProfiler:  "process_id:31145,service:zy-profiling-test,profiler_version:0.102.0~b67f6e3380,host:zydeMacBook-Air.local,runtime-id:06dddda1-957b-4619-97cb-1a78fc7e3f07,language:jvm,env:test,version:v1.2",
		SubCustomTags: "foobar:hello-world",
		Start:         metrics.NewRFC3339Time(time.Now()),
		End:           metrics.NewRFC3339Time(time.Now().Add(time.Minute)),
	}

	out, err := json.MarshalIndent(md, "", "    ")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(out))

	var md2 Metadata

	if err = json.Unmarshal(out, &md2); err != nil {
		t.Fatal(err)
	}

	fmt.Println("start: ", time.Time(*md2.Start))
	fmt.Println("end: ", time.Time(*md2.End))

	var m map[string]any

	if err = json.Unmarshal(out, &m); err != nil {
		t.Fatal(err)
	}

	headers := json2FormValues(m)

	for k, v := range headers {
		fmt.Println(k, ":", v)
	}
}

func TestJson2StringMap(t *testing.T) {
	jsonStr := `
{
    "attachments": [
        "main.jfr",
        "metrics.json"
    ],
    "tags_profiler": "process_id:31145,service:zy-profiling-test,profiler_version:0.102.0~b67f6e3380,host:zydeMacBook-Air.local,runtime-id:06dddda1-957b-4619-97cb-1a78fc7e3f07,language:jvm,env:test,version:v1.2",
    "start": "2022-06-17T09:20:07.002305Z",
    "end": "2022-06-17T09:21:08.261768Z",
    "family": "java",
    "version": "4",
	"numbers": [1, 3, 5],
	"stable": false
}
`

	var v map[string]interface{}

	err := json.Unmarshal([]byte(jsonStr), &v)
	if err != nil {
		t.Fatal(err)
	}

	strMap := json2FormValues(v)

	for key, val := range strMap {
		fmt.Println(key, ":", val)
	}
}

var eventJSON = `
{
  "version": "4",
  "family": "python",
  "attachments": [
    "auto.pprof",
    "code-provenance.json"
  ],
  "tags_profiler": "service:python-profiling-demo,runtime-id:1f59e0c9a247437a966e4f4e3375de8e,foo:bar,foobar:hello-world,host:SpaceX.local,language:python,runtime:CPython,runtime_version:3.10.5,profiler_version:1.17.0,version:v0.0.1,env:testing",
  "start": "2023-08-01T13:07:03Z",
  "end": "2023-08-01T13:07:05Z",
  "endpoint_counts": {}
}
`

func TestParseTags(t *testing.T) {
	req, err := http.NewRequest("POST", "/ingest?aggregationType=sum&from=1734427928674085000&name=go-pyroscope-demo{__session_id__=718f0dfa2700ee16,env=demo,host=SpaceX.local,service=go-pyroscope-demo,version=0.0.1}&sampleRate=100&spyName=gospy&units=samples&until=1734427988708602000", nil)
	if err != nil {
		t.Fatal(err)
	}

	metadata := ParseTags(QueryToPBForms(req.URL.Query()))

	for k, v := range metadata {
		t.Logf("%s : %s \n", k, v)
	}

	fmt.Println(metrics.JoinTags(metadata))
}
