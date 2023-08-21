// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package profile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/GuanceCloud/cliutils/testutil"
)

func TestJson2StringMap(t *testing.T) {
	jsonStr := `{"attachments":["main.jfr"],"tags_profiler":"process_id:31145,service:zy-profiling-test,profiler_version:0.102.0~b67f6e3380,host:zydeMacBook-Air.local,runtime-id:06dddda1-957b-4619-97cb-1a78fc7e3f07,language:jvm,env:test,version:v1.2","start":"2022-06-17T09:20:07.002305Z","end":"2022-06-17T09:21:08.261768Z","family":"java","version":"4"}`

	var v map[string]interface{}

	err := json.Unmarshal([]byte(jsonStr), &v)
	if err != nil {
		t.Fatal(err)
	}

	strMap := json2StringMap(v)

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

func TestParseMetadata(t *testing.T) {
	var buf bytes.Buffer

	w := multipart.NewWriter(&buf)

	f, err := w.CreateFormFile("event", "event.json")
	testutil.Ok(t, err)

	_, err = f.Write([]byte(eventJSON))
	testutil.Ok(t, err)

	err = w.Close()
	testutil.Ok(t, err)

	req, err := http.NewRequest("POST", "/profiling/v1/input", &buf)
	testutil.Ok(t, err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	metadata, _, err := parseMetadata(req)

	testutil.Ok(t, err)

	for k, v := range metadata.tags {
		t.Logf("%s : %s \n", k, v)
	}

	testutil.Equals(t, "bar", metadata.tags["foo"])
	testutil.Equals(t, "hello-world", metadata.tags["foobar"])
}
