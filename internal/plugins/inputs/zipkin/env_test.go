// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package zipkin handle Zipkin APM traces.
package zipkin

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/workerpool"
)

func loadEnvs() map[string]string {
	envs := make(map[string]string)
	for _, env := range os.Environ() {
		ss := strings.Split(env, "=")
		if len(ss) == 2 {
			envs[ss[0]] = ss[1]
		}
	}

	return envs
}

func TestReadEnv(t *testing.T) {
	cases := []struct {
		name     string
		envs     map[string]string
		expected *Input
	}{
		{
			name: "zip_env_tc_1",
			envs: map[string]string{
				"ENV_INPUT_ZIPKIN_PATH_V1":            "/api/v1/spans",
				"ENV_INPUT_ZIPKIN_PATH_V2":            "/api/v2/spans",
				"ENV_INPUT_ZIPKIN_IGNORE_TAGS":        `["block1", "block2"]`,
				"ENV_INPUT_ZIPKIN_KEEP_RARE_RESOURCE": "true",
				"ENV_INPUT_ZIPKIN_CLOSE_RESOURCE":     `{"service1":["resource1"], "service2":["resource2"], "service3":["resource3"]}`,
				"ENV_INPUT_ZIPKIN_SAMPLER":            "0.3",
				"ENV_INPUT_ZIPKIN_TAGS":               `{"k1":"v1", "k2":"v2", "k3":"v3"}`,
				"ENV_INPUT_ZIPKIN_THREADS":            `{"buffer":1000, "threads":100}`,
				"ENV_INPUT_ZIPKIN_STORAGE":            `{"storage":"./zipkin_storage", "capacity": 5120}`,
			},
			expected: &Input{
				PathV1:           "/api/v1/spans",
				PathV2:           "/api/v2/spans",
				IgnoreTags:       []string{"block1", "block2"},
				KeepRareResource: true,
				CloseResource:    map[string][]string{"service1": {"resource1"}, "service2": {"resource2"}, "service3": {"resource3"}},
				Sampler:          &trace.Sampler{SamplingRateGlobal: 0.3},
				Tags:             map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"},
				WPConfig:         &workerpool.WorkerPoolConfig{Buffer: 1000, Threads: 100},
				LocalCacheConfig: &storage.StorageConfig{Path: "./zipkin_storage", Capacity: 5120},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// set env values
			for k, v := range tc.envs {
				assert.NoError(t, os.Setenv(k, v))
			}
			// call ddtrace.Input.ReadEnv
			ipt := &Input{}
			ipt.ReadEnv(loadEnvs())
			// compare
			assert.Equal(t, tc.expected, ipt, "test case: %s, expected: %v, actual: %v", tc.expected, ipt)
			// clear env values
			t.Cleanup(func() {
				for k := range tc.envs {
					assert.NoError(t, os.Unsetenv(k))
				}
			})
		})
	}
}
