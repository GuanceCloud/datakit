// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package jit

import (
	"os"
	"reflect"
	"testing"
)

var userAgentOutputKeys = []string{
	"isMobile", "isBot", "os", "browser", "browserVer", "engine", "engineVer", "ua",
}

func TestNativeUserAgentEndToEnd(t *testing.T) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		t.Skip("set PLATYPUS_JIT_RUNTIME to a trusted platypus_jit cdylib")
	}
	runner, err := NewRunnerWithHost(runtimePath, "pipeline-go-1.4.3-datakit", 8, NewPipelineGoHost(nil))
	if err != nil {
		t.Fatalf("open runner with compatibility host: %v", err)
	}
	defer runner.Close()

	check := runner.Check("user_agent(agent)\n")
	if check.Route != RouteJITNative || check.Capabilities.RequiredHostFlags != 0 ||
		len(check.Capabilities.HostCalls) != 0 || check.Capabilities.ExecutionMode != "native" {
		t.Fatalf("user_agent must be purely native: %#v", check)
	}

	tests := []struct {
		name  string
		input string
		want  map[string]any
	}{
		{
			name:  "chrome",
			input: "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.72 Safari/537.36",
			want: map[string]any{
				"isMobile": false, "isBot": false, "os": "Intel Mac OS X 11_1_0",
				"browser": "Chrome", "browserVer": "89.0.4389.72",
				"engine": "AppleWebKit", "engineVer": "537.36", "ua": "Macintosh",
			},
		},
		{
			name:  "iphone",
			input: "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 Version/17.0 Mobile/15E148 Safari/604.1",
			want: map[string]any{
				"isMobile": true, "isBot": false, "os": "CPU iPhone OS 17_0 like Mac OS X",
				"browser": "Safari", "browserVer": "17.0",
				"engine": "AppleWebKit", "engineVer": "605.1.15", "ua": "iPhone",
			},
		},
		{
			name:  "curl",
			input: "curl/8.0",
			want: map[string]any{
				"isMobile": false, "isBot": false, "os": "",
				"browser": "curl", "browserVer": "8.0",
				"engine": "", "engineVer": "", "ua": "",
			},
		},
		{
			name:  "okhttp",
			input: "okhttp/4.2.2",
			want: map[string]any{
				"isMobile": true, "isBot": false, "os": "",
				"browser": "OkHttp", "browserVer": "4.2.2",
				"engine": "", "engineVer": "", "ua": "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			point := runNativeUserAgentPoint(t, runner, "user_agent(agent)\n", Point{
				Version: 1, Category: "logging", Measurement: test.name,
				Fields: map[string]any{"agent": test.input},
			})
			assertUserAgentValues(t, point, "", test.want)
		})
	}

	t.Run("missing-key-is-noop", func(t *testing.T) {
		point := runNativeUserAgentPoint(t, runner, "user_agent(agent)\n", Point{
			Version: 1, Category: "logging", Measurement: "missing",
			Fields: map[string]any{"sentinel": "keep"},
		})
		if point.Fields["sentinel"] != "keep" {
			t.Fatalf("missing key changed sentinel: %#v", point)
		}
		for _, key := range userAgentOutputKeys {
			if _, field := point.Fields[key]; field {
				t.Fatalf("missing key created field %q: %#v", key, point)
			}
			if _, tag := point.Tags[key]; tag {
				t.Fatalf("missing key created tag %q: %#v", key, point)
			}
		}
	})

	t.Run("prefix", func(t *testing.T) {
		want := tests[2].want
		point := runNativeUserAgentPoint(t, runner, "user_agent(agent, \"ua_\")\n", Point{
			Version: 1, Category: "logging", Measurement: "prefix",
			Fields: map[string]any{"agent": tests[2].input, "browser": "keep"},
		})
		assertUserAgentValues(t, point, "ua_", want)
		if point.Fields["browser"] != "keep" {
			t.Fatalf("prefixed user_agent overwrote unprefixed browser: %#v", point)
		}
	})

	t.Run("existing-fields-are-overwritten", func(t *testing.T) {
		want := tests[0].want
		point := runNativeUserAgentPoint(t, runner, "user_agent(agent)\n", Point{
			Version: 1, Category: "logging", Measurement: "field-overwrite",
			Fields: map[string]any{
				"agent": tests[0].input, "browser": "old", "isMobile": "old",
			},
		})
		assertUserAgentValues(t, point, "", want)
	})

	t.Run("existing-tags-remain-tags", func(t *testing.T) {
		want := tests[0].want
		point := runNativeUserAgentPoint(t, runner, "user_agent(agent)\n", Point{
			Version: 1, Category: "logging", Measurement: "tag-overwrite",
			Tags:   map[string]string{"browser": "old", "isMobile": "old"},
			Fields: map[string]any{"agent": tests[0].input},
		})
		if point.Tags["browser"] != "Chrome" || point.Tags["isMobile"] != "false" {
			t.Fatalf("existing tags were not overwritten as tags: %#v", point)
		}
		if _, ok := point.Fields["browser"]; ok {
			t.Fatalf("browser tag was converted to a field: %#v", point)
		}
		if _, ok := point.Fields["isMobile"]; ok {
			t.Fatalf("isMobile tag was converted to a field: %#v", point)
		}
		for key, value := range want {
			if key == "browser" || key == "isMobile" {
				continue
			}
			if !reflect.DeepEqual(point.Fields[key], value) {
				t.Fatalf("field %q = %#v, want %#v; point=%#v", key, point.Fields[key], value, point)
			}
		}
	})
}

func assertUserAgentValues(t *testing.T, point Point, prefix string, want map[string]any) {
	t.Helper()
	for key, value := range want {
		if !reflect.DeepEqual(point.Fields[prefix+key], value) {
			t.Fatalf("field %q = %#v, want %#v; point=%#v", prefix+key, point.Fields[prefix+key], value, point)
		}
	}
}

func runNativeUserAgentPoint(t *testing.T, runner *Runner, source string, point Point) Point {
	t.Helper()
	input, err := EncodeFlatPoints([]Point{point})
	if err != nil {
		t.Fatalf("encode user-agent point: %v", err)
	}
	batch, err := runner.Process(source, input)
	if err != nil {
		t.Fatalf("process user-agent point: %v", err)
	}
	if len(batch.Records) != 1 || batch.Records[0].Status != TerminalOK {
		t.Fatalf("unexpected user-agent batch: %#v", batch)
	}
	if batch.Static == nil {
		deltas, err := batch.Records[0].MutationDeltas()
		if err != nil || len(deltas) != 1 {
			t.Fatalf("decode user-agent delta: count=%d err=%v", len(deltas), err)
		}
		if err := deltas[0].Apply(&point); err != nil {
			t.Fatalf("apply user-agent delta: %v", err)
		}
		return point
	}

	static := batch.Static
	if len(static.StateOffsets) != 2 || len(static.ValueOffsets) != 2 ||
		int(static.StateOffsets[1]) != len(static.Schema.Keys) {
		t.Fatalf("invalid user-agent static batch: %#v", static)
	}
	valueIndex := int(static.ValueOffsets[0])
	valueEnd := int(static.ValueOffsets[1])
	for index, key := range static.Schema.Keys {
		state := static.States[index]
		switch state {
		case StaticNoop:
		case StaticDeleteField, StaticDeleteTag:
			delete(point.Fields, key)
			delete(point.Tags, key)
		case StaticSetField:
			if valueIndex >= valueEnd {
				t.Fatalf("static field %q is missing its value", key)
			}
			if point.Fields == nil {
				point.Fields = map[string]any{}
			}
			delete(point.Tags, key)
			point.Fields[key] = static.Values[valueIndex]
			valueIndex++
		case StaticSetTag:
			if valueIndex >= valueEnd {
				t.Fatalf("static tag %q is missing its value", key)
			}
			value, ok := static.Values[valueIndex].(string)
			if !ok {
				t.Fatalf("static tag %q has value type %T", key, static.Values[valueIndex])
			}
			if point.Tags == nil {
				point.Tags = map[string]string{}
			}
			delete(point.Fields, key)
			point.Tags[key] = value
			valueIndex++
		default:
			t.Fatalf("static key %q has unknown state %d", key, state)
		}
	}
	if valueIndex != valueEnd {
		t.Fatalf("static user-agent batch has %d unused values", valueEnd-valueIndex)
	}
	return point
}
