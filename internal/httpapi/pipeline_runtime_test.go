// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPipelineRuntimeGrokAndJSON(t *testing.T) {
	cases := []struct {
		name       string
		scriptName string
		script     string
		data       string
		assertions map[string]string
	}{
		{
			name:       "grok access log",
			scriptName: "access",
			script: `
grok(_, "%{IPORHOST:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{WORD:http_method} %{URIPATHPARAM:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")
cast(status_code, "int")
cast(bytes, "int")
`,
			data: `10.20.30.40 - - [13/May/2026:10:11:12 +0800] "GET /api/v1/pipeline?q=json HTTP/1.1" 200 512`,
			assertions: map[string]string{
				"client_ip":    "10.20.30.40",
				"http_method":  "GET",
				"http_url":     "/api/v1/pipeline?q=json",
				"http_version": "1.1",
				"status_code":  "200",
				"bytes":        "512",
			},
		},
		{
			name:       "json log",
			scriptName: "jsonlog",
			script: `
json(_, service, "service")
json(_, status, "json_status")
json(_, cost, "cost")
json(_, message, "msg")
cast(cost, "float")
`,
			data: `{"service":"datakit","status":"ok","cost":12.75,"message":"json pipeline runtime"}`,
			assertions: map[string]string{
				"service":     "datakit",
				"json_status": "ok",
				"cost":        "12.75",
				"msg":         "json pipeline runtime",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := &pipelineDebugRequest{
				Pipeline: map[string]map[string]string{
					"logging": {
						tc.scriptName: base64.StdEncoding.EncodeToString([]byte(tc.script)),
					},
				},
				Category:   "logging",
				ScriptName: tc.scriptName,
				DataType:   "text/plain",
				Data:       []string{base64.StdEncoding.EncodeToString([]byte(tc.data))},
				Benchmark:  true,
			}
			body, err := json.Marshal(reqBody)
			if err != nil {
				t.Fatal(err)
			}

			resp, err := apiPipelineDebugHandler(
				httptest.NewRecorder(),
				httptest.NewRequest("POST", "/v1/pipeline", bytes.NewReader(body)),
			)
			if err != nil {
				t.Fatal(err)
			}
			result, ok := resp.(*pipelineDebugResponse)
			if !ok {
				t.Fatalf("unexpected response type %T", resp)
			}
			if len(result.PlErrors) > 0 {
				t.Fatalf("pipeline parse errors: %+v", result.PlErrors)
			}
			if len(result.PLResults) != 1 || result.PLResults[0].RunError != nil {
				t.Fatalf("pipeline run failed: %+v", result.PLResults)
			}
			if result.Benchmark == "" || !strings.Contains(result.Benchmark, "ns/op") {
				t.Fatalf("missing benchmark result: %q", result.Benchmark)
			}

			fields := result.PLResults[0].Point.Fields
			for k, want := range tc.assertions {
				if got := fmt.Sprint(fields[k]); got != want {
					t.Fatalf("field %q = %q, want %q; all fields: %#v", k, got, want, fields)
				}
			}
			t.Logf("%s benchmark: %s", tc.name, result.Benchmark)
		})
	}
}
