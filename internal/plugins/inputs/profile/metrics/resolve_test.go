// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package metrics

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetadata(t *testing.T) {
	md := &Metadata{
		Format:        Collapsed,
		Profiler:      Pyroscope,
		Attachments:   []string{"main.jfr", "metrics.json"},
		Language:      Golang,
		TagsProfiler:  "process_id:31145,service:zy-profiling-test,profiler_version:0.102.0~b67f6e3380,host:zydeMacBook-Air.local,runtime-id:06dddda1-957b-4619-97cb-1a78fc7e3f07,language:jvm,env:test,version:v1.2",
		SubCustomTags: "foobar:hello-world",
		Start:         NewRFC3339Time(time.Now()),
		End:           NewRFC3339Time(time.Now().Add(time.Minute)),
	}

	out, err := json.MarshalIndent(md, "", "    ")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(out))

	var md2 Metadata

	if err = json.Unmarshal(out, &md2); err != nil {
		t.Fatal(err)
	}

	assert.True(t, md.Start.Before(md.End))
	assert.True(t, md.End.After(md.Start))

	t.Log("start: ", time.Time(*md2.Start))
	t.Log("end: ", time.Time(*md2.End))

	var m map[string]any

	if err = json.Unmarshal(out, &m); err != nil {
		t.Fatal(err)
	}

	headers := json2FormValues(m)

	for k, v := range headers {
		t.Log(k, ":", v)
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
		t.Log(key, ":", val)
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

func TestParseMetadata(t *testing.T) {
	var buf bytes.Buffer

	w := multipart.NewWriter(&buf)

	f, err := w.CreateFormFile("event", "event.json")
	assert.NoError(t, err)

	_, err = f.Write([]byte(eventJSON))
	assert.NoError(t, err)

	err = w.Close()
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/profiling/v1/input", &buf)
	assert.NoError(t, err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	err = req.ParseMultipartForm(1e9)
	assert.NoError(t, err)

	metadata, _, err := ParseMetadata(req)

	assert.NoError(t, err)

	for k, v := range metadata {
		t.Logf("%s : %s \n", k, v)
	}

	assert.Equal(t, "bar", metadata["foo"])
	assert.Equal(t, "hello-world", metadata["foobar"])

	t.Log(JoinTags(metadata))
}
