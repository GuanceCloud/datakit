// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package httpapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

func TestPipelineDebugGrokFastPathSwitch(t *testing.T) {
	t.Cleanup(func() {
		if err := plval.InitPlVal(&plval.PipelineCfg{EnableGrokFastPath: false}, nil, nil, ""); err != nil {
			t.Fatal(err)
		}
		funcs.SetGrokRunObserver(nil)
	})

	run := func(enabled bool) funcs.GrokRunInfo {
		t.Helper()

		if err := plval.InitPlVal(&plval.PipelineCfg{EnableGrokFastPath: enabled}, nil, nil, ""); err != nil {
			t.Fatal(err)
		}

		var got funcs.GrokRunInfo
		funcs.SetGrokRunObserver(func(info funcs.GrokRunInfo) {
			got = info
		})

		reqBody := &pipelineDebugRequest{
			Pipeline: map[string]map[string]string{
				"logging": {
					"access": base64.StdEncoding.EncodeToString([]byte(`
grok(_, "%{IPORHOST:client_ip} - - \\[%{HTTPDATE:time}\\] \"%{WORD:http_method} %{URIPATHPARAM:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")
`)),
				},
			},
			Category:   "logging",
			ScriptName: "access",
			DataType:   "text/plain",
			Data: []string{
				base64.StdEncoding.EncodeToString([]byte(
					`10.20.30.40 - - [13/May/2026:10:11:12 +0800] "GET /api/v1/pipeline?q=json HTTP/1.1" 200 512`,
				)),
			},
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
		if got.Path == "" {
			t.Fatal("missing grok run observer result")
		}

		return got
	}

	disabled := run(false)
	if disabled.Path != "regexp" || disabled.FallbackReason != "fast_path_disabled" {
		t.Fatalf("disabled path = %q/%q, want regexp/fast_path_disabled", disabled.Path, disabled.FallbackReason)
	}

	enabled := run(true)
	if enabled.Path != "fast_path" || enabled.FallbackReason != "none" {
		t.Fatalf("enabled path = %q/%q, want fast_path/none", enabled.Path, enabled.FallbackReason)
	}
}
