// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package skywalking handle SkyWalking tracing metrics.
package skywalking

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
			name: "sky_env_tc_1",
			envs: map[string]string{
				"ENV_INPUT_SKYWALKING_HTTP_ENDPOINTS":     `["/v3/trace", "/v3/metric", "/v3/logging", "/v3/profiling"]`,
				"ENV_INPUT_SKYWALKING_GRPC_ENDPOINT":      "127.0.0.1:11800",
				"ENV_INPUT_SKYWALKING_PLUGINS":            `["db.type", "os.call"]`,
				"ENV_INPUT_SKYWALKING_CUSTOMER_TAGS":      `["key1", "key2", "key3"]`,
				"ENV_INPUT_SKYWALKING_KEEP_RARE_RESOURCE": "true",
				"ENV_INPUT_SKYWALKING_CLOSE_RESOURCE":     `{"service1":["resource1"], "service2":["resource2"], "service3":["resource3"]}`,
				"ENV_INPUT_SKYWALKING_SAMPLER":            "0.3",
				"ENV_INPUT_SKYWALKING_TAGS":               `{"k1":"v1", "k2":"v2", "k3":"v3"}`,
				"ENV_INPUT_SKYWALKING_THREADS":            `{"buffer":1000, "threads":100}`,
				"ENV_INPUT_SKYWALKING_STORAGE":            `{"storage":"./skywalking_storage", "capacity": 5120}`,
			},
			expected: &Input{
				Endpoints:        []string{"/v3/trace", "/v3/metric", "/v3/logging", "/v3/profiling"},
				Address:          "127.0.0.1:11800",
				Plugins:          []string{"db.type", "os.call"},
				CustomerTags:     []string{"key1", "key2", "key3"},
				KeepRareResource: true,
				CloseResource:    map[string][]string{"service1": {"resource1"}, "service2": {"resource2"}, "service3": {"resource3"}},
				Sampler:          &trace.Sampler{SamplingRateGlobal: 0.3},
				Tags:             map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"},
				WPConfig:         &workerpool.WorkerPoolConfig{Buffer: 1000, Threads: 100},
				LocalCacheConfig: &storage.StorageConfig{Path: "./skywalking_storage", Capacity: 5120},
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
				os.Clearenv()
			})
		})
	}
}
