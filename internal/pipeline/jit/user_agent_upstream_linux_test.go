// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jit

import (
	"os"
	"testing"
)

func TestUserAgentPipelineGoUpstreamCorpus(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 24)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	windows := "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1985.125 Safari/537.36"
	mac := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.1 Safari/605.1.15"
	windowsValues := map[string]any{
		"isMobile": false, "isBot": false, "os": "Windows 7", "browser": "Chrome",
		"browserVer": "36.0.1985.125", "engine": "AppleWebKit", "engineVer": "537.36", "ua": "Windows",
	}
	macValues := map[string]any{
		"isMobile": false, "isBot": false, "os": "Intel Mac OS X 10_15_7", "browser": "Safari",
		"browserVer": "15.1", "engine": "AppleWebKit", "engineVer": "605.1.15", "ua": "Macintosh",
	}
	cases := []struct {
		name     string
		source   string
		message  string
		prefix   string
		expected map[string]any
		preserve map[string]any
	}{
		{"upstream-windows", `json(_,userAgent); user_agent(userAgent); add_key(after,true)`, `{"userAgent":` + quoteJSON(windows) + `,"second":2,"third":"abc","forth":true}`, "", windowsValues, nil},
		{"upstream-mac", `json(_,userAgent); user_agent(userAgent); add_key(after,true)`, `{"userAgent":` + quoteJSON(mac) + `}`, "", macValues, nil},
		{"upstream-missing-key", `json(_,userAgent); user_agent(agent); add_key(after,true)`, `{"userAgent":` + quoteJSON(mac) + `}`, "", nil, nil},
		// getKeyName accepts a string literal and resolves its contents as the Point key.
		{"upstream-string-key", `json(_,userAgent); user_agent("userAgent"); add_key(after,true)`, `{"userAgent":` + quoteJSON(mac) + `}`, "", macValues, nil},
		{"upstream-prefix", `json(_,userAgent); add_key(os,"existing"); user_agent(userAgent,"ua_"); add_key(after,true)`, `{"userAgent":` + quoteJSON(windows) + `}`, "ua_", windowsValues, map[string]any{"os": "existing"}},
		{"upstream-variable-prefix", `json(_,userAgent); p="ua_"; user_agent(userAgent,p); add_key(after,true)`, `{"userAgent":` + quoteJSON(windows) + `}`, "ua_", windowsValues, nil},
		{"upstream-named-prefix", `json(_,userAgent); user_agent(userAgent,prefix="ua_"); add_key(after,true)`, `{"userAgent":` + quoteJSON(windows) + `}`, "ua_", windowsValues, nil},
		{"upstream-empty-prefix", `json(_,userAgent); user_agent(userAgent,""); add_key(after,true)`, `{"userAgent":` + quoteJSON(windows) + `}`, "", windowsValues, nil},
		{"upstream-missing-with-prefix", `user_agent(agent,"ua_"); add_key(after,true)`, `{}`, "ua_", nil, nil},
	}

	for _, size := range []int{1, 2, 4, 8, 10, 128} {
		t.Run("batch-"+itoa(size), func(t *testing.T) {
			for _, tc := range cases {
				check := runner.Check(tc.source)
				if check.Route != RouteJITNative || check.Capabilities.ExecutionMode != "native" ||
					check.Capabilities.RequiredHostFlags != 0 || len(check.Capabilities.HostCalls) != 0 {
					t.Fatalf("%s route=%+v", tc.name, check)
				}
				points := make([]Point, size)
				for i := range points {
					points[i] = Point{Version: 1, Category: "logging", Measurement: tc.name,
						Fields: map[string]any{"message": tc.message, "sequence": int64(i)}}
				}
				input, err := EncodeFlatPoints(points)
				if err != nil {
					t.Fatal(err)
				}
				batch, err := runner.Process(tc.source, input)
				if err != nil || len(batch.Records) != size {
					t.Fatalf("%s records=%d error=%v", tc.name, len(batch.Records), err)
				}
				for i, record := range batch.Records {
					if record.Status != TerminalOK {
						t.Fatalf("%s record %d terminal=%d error=%s", tc.name, i, record.Status, record.Error)
					}
					applyHostCompatRecord(t, batch, i, &points[i])
					for key, want := range tc.expected {
						if got := points[i].Fields[tc.prefix+key]; got != want {
							t.Fatalf("%s record %d %s=%#v want %#v", tc.name, i, tc.prefix+key, got, want)
						}
					}
					if tc.expected == nil {
						for _, key := range userAgentOutputKeys {
							if _, ok := points[i].Fields[tc.prefix+key]; ok {
								t.Fatalf("%s record %d unexpectedly set %s: %#v", tc.name, i, tc.prefix+key, points[i])
							}
						}
					}
					for key, want := range tc.preserve {
						if points[i].Fields[key] != want {
							t.Fatalf("%s record %d did not preserve %s: %#v", tc.name, i, key, points[i])
						}
					}
					if points[i].Fields["after"] != true || points[i].Fields["sequence"] != int64(i) {
						t.Fatalf("%s record %d lost continuation/input: %#v", tc.name, i, points[i])
					}
				}
			}
		})
	}

	for _, source := range []string{
		`json(_,userAgent); user_agent(userAgent,123)`,
		`json(_,userAgent); user_agent(userAgent,someArg,anotherArg)`,
	} {
		check := runner.Check(source)
		if check.Route != RoutePipelineGo || check.Reason != CheckReasonCompileError {
			t.Fatalf("invalid script admitted: source=%q check=%+v", source, check)
		}
	}

	t.Run("dynamic-non-string-prefix-terminal", func(t *testing.T) {
		const source = `json(_,userAgent); add_key(before,true); p=123; user_agent(userAgent,p); add_key(after,true)`
		check := runner.Check(source)
		if check.Route != RouteJITNative {
			t.Fatalf("route=%+v", check)
		}
		point := Point{Version: 1, Category: "logging", Measurement: "invalid-prefix",
			Fields: map[string]any{"message": `{"userAgent":` + quoteJSON(windows) + `}`}}
		input, err := EncodeFlatPoints([]Point{point})
		if err != nil {
			t.Fatal(err)
		}
		batch, err := runner.Process(source, input)
		if err != nil || len(batch.Records) != 1 || batch.Records[0].Status != TerminalError || !batch.Records[0].CommitPrefixError {
			t.Fatalf("batch=%+v error=%v", batch, err)
		}
		applyHostCompatRecord(t, batch, 0, &point)
		if point.Fields["before"] != true {
			t.Fatalf("prefix not committed: %#v", point)
		}
		if _, ok := point.Fields["after"]; ok {
			t.Fatalf("terminal error continued: %#v", point)
		}
	})
}

func quoteJSON(value string) string {
	// UA fixtures contain only JSON-safe ASCII; quoting here keeps the table readable.
	return `"` + value + `"`
}
