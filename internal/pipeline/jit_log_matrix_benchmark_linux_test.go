// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && jitbench && (amd64 || arm64)

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func loadLogMatrixWorkloads(b testing.TB) []realScriptBenchmarkWorkload {
	b.Helper()
	root := os.Getenv("JIT_BENCH_LOG_DIRECTORY")
	if root == "" {
		root = filepath.Join("testdata", "jit-log-matrix")
	}
	read := func(name string) []byte {
		b.Helper()
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			b.Fatal(err)
		}
		return data
	}
	var cases []struct {
		Name               string           `json:"name"`
		Script             string           `json:"script"`
		Input              string           `json:"input"`
		Expected           map[string]any   `json:"expected"`
		ExtraFields        map[string]any   `json:"extra_fields"`
		ExtraIntegerFields map[string]int64 `json:"extra_integer_fields"`
	}
	if err := json.Unmarshal(read("cases.json"), &cases); err != nil {
		b.Fatal(err)
	}
	if len(cases) == 0 {
		b.Fatal("empty log matrix")
	}
	workloads := make([]realScriptBenchmarkWorkload, 0, len(cases))
	for _, c := range cases {
		if len(c.Expected) == 0 {
			b.Fatalf("%s has no required output", c.Name)
		}
		if c.ExtraFields == nil {
			c.ExtraFields = make(map[string]any)
		}
		for key, value := range c.ExtraIntegerFields {
			c.ExtraFields[key] = value
		}
		workloads = append(workloads, realScriptBenchmarkWorkload{
			name: c.Name, script: c.Script, source: string(read(c.Script)), message: string(read(c.Input)),
			expectedFields: c.Expected, extraFields: c.ExtraFields,
		})
	}
	return workloads
}
