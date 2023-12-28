// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package opentelemetry handle OTEL APM trace
package opentelemetry

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
			name: "otel_env_tc_1",
			envs: map[string]string{
				"ENV_INPUT_OTEL_CUSTOMER_TAGS":      `["block1", "block2"]`,
				"ENV_INPUT_OTEL_KEEP_RARE_RESOURCE": "true",
				"ENV_INPUT_OTEL_OMIT_ERR_STATUS":    `["404", "403", "400"]`,
				"ENV_INPUT_OTEL_CLOSE_RESOURCE":     `{"service1":["resource1"], "service2":["resource2"], "service3":["resource3"]}`,
				"ENV_INPUT_OTEL_SAMPLER":            "0.3",
				"ENV_INPUT_OTEL_TAGS":               `{"k1":"v1", "k2":"v2", "k3":"v3"}`,
				"ENV_INPUT_OTEL_THREADS":            `{"buffer":1000, "threads":100}`,
				"ENV_INPUT_OTEL_STORAGE":            `{"storage":"./otel_storage", "capacity": 5120}`,
				"ENV_INPUT_OTEL_HTTP":               `{"enable":true, "http_status_ok": 200, "trace_api": "/otel/v1/trace", "metric_api": "/otel/v1/metric"}`,
				"ENV_INPUT_OTEL_GRPC":               `{"trace_enable": true, "metric_enable": true, "addr": "127.0.0.1:4317"}`,
				"ENV_INPUT_OTEL_EXPECTED_HEADERS":   `{"ex_version": "1.2.3", "ex_name": "env_resource_name"}`,
			},
			expected: &Input{
				CustomerTags:     []string{"block1", "block2"},
				KeepRareResource: true,
				OmitErrStatus:    []string{"404", "403", "400"},
				CloseResource:    map[string][]string{"service1": {"resource1"}, "service2": {"resource2"}, "service3": {"resource3"}},
				Sampler:          &trace.Sampler{SamplingRateGlobal: 0.3},
				Tags:             map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"},
				WPConfig:         &workerpool.WorkerPoolConfig{Buffer: 1000, Threads: 100},
				LocalCacheConfig: &storage.StorageConfig{Path: "./otel_storage", Capacity: 5120},
				HTTPConfig:       &httpConfig{Enabled: true, StatusCodeOK: 200, TraceAPI: "/otel/v1/trace", MetricAPI: "/otel/v1/metric"},
				GRPCConfig:       &grpcConfig{TraceEnabled: true, MetricEnabled: true, Address: "127.0.0.1:4317"},
				ExpectedHeaders:  map[string]string{"ex_version": "1.2.3", "ex_name": "env_resource_name"},
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
