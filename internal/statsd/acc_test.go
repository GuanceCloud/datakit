// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package statsd

import (
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	tu "github.com/GuanceCloud/cliutils/testutil"
	"github.com/stretchr/testify/require"
)

func TestAddFields(t *testing.T) {
	cases := []struct {
		name   string
		tags   map[string]string
		fields map[string]interface{}

		mmap        []string
		dropTags    []string
		expectPoint int
	}{
		{
			mmap:     []string{"jvm_:jvm"},
			dropTags: []string{"c"},

			name:        `jvm_cpu_load_process`,
			tags:        map[string]string{"a": "b", "c": "d"},
			fields:      map[string]interface{}{"value": 1024},
			expectPoint: 1,
		},

		{
			mmap:     nil,
			dropTags: []string{"c"},

			name:        `jvm_cpu_load_process`,
			tags:        map[string]string{"a": "b", "c": "d"},
			fields:      map[string]interface{}{"value": 1024},
			expectPoint: 1,
		},

		{
			mmap:     nil,
			dropTags: []string{"c"},

			// warning name, no `_'(the default) seprator, we choose accept it
			name:        `jvmcpuloadprocess`,
			tags:        map[string]string{"a": "b", "c": "d"},
			fields:      map[string]interface{}{"value": 1024},
			expectPoint: 1,
		},

		{
			mmap:     []string{"net:set"},
			dropTags: []string{"c"},

			// warning name, no `_'(the default) seprator, we choose accept it
			name:        `dotnet_cpu_load_process`,
			tags:        map[string]string{"a": "b", "c": "d"},
			fields:      map[string]interface{}{"value": 1024},
			expectPoint: 1,
		},

		{
			mmap:     []string{"jvm_cpu_:jvmcpu"},
			dropTags: []string{"c"},

			name:        `jvm_cpu_load_process`,
			tags:        map[string]string{"a": "b", "c": "d"},
			fields:      map[string]interface{}{"value": 1024},
			expectPoint: 1,
		},

		{
			mmap:     []string{"jvm_:jvm"},
			dropTags: []string{"c"},

			name:        `jvm_cpu_load_process`,
			tags:        map[string]string{"a": "b", "c": "d"},
			fields:      map[string]interface{}{"invalid-field": 1024},
			expectPoint: 0,
		},

		{
			mmap:     []string{"jvm_:jvm"},
			dropTags: []string{"c"},

			name:        `jvm_cpu_load_process`,
			tags:        map[string]string{"a": "b", "c": "d"},
			fields:      map[string]interface{}{"invalid-field": 1024, "field": 42},
			expectPoint: 0,
		},
	}

	opt := option{}
	s := &Collector{opts: &opt}
	acc := &accumulator{
		ref: s,
		l:   logger.SLogger("ioName"),
	}
	s.acc = acc

	for _, tc := range cases {
		acc.points = acc.points[:0] // clear cache

		s.opts.metricMapping = tc.mmap
		s.opts.dropTags = tc.dropTags
		s.setupMmap()

		acc.addFields(tc.name, tc.fields, tc.tags, time.Now())

		tu.Assert(t, len(acc.points) == tc.expectPoint,
			"expect %d point, got %d: %+#v",
			tc.expectPoint, len(acc.points), acc.points)

		if len(acc.points) > 0 {
			t.Logf("%#v", acc.points[len(acc.points)-1])
		}
	}
}

// go test -v -timeout 30s -run ^TestDoFeedMetricName$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/statsd
func TestDoFeedMetricName(t *testing.T) {
	cases := []struct {
		name                 string
		acc                  *accumulator
		tags                 map[string]string
		expectFeedMetricName string
	}{
		{
			name: "normal",
			acc: &accumulator{
				ref: &Collector{
					opts: &option{
						statsdSourceKey: "source_key",
						statsdHostKey:   "host_key",
					},
				},
			},
			tags: map[string]string{
				"source_key": "tomcat",
				"host_key":   "cn-shanghai-sq5ei",
			},
			expectFeedMetricName: "statsd/tomcat/cn-shanghai-sq5ei",
		},

		{
			name: "default",
			acc: &accumulator{
				ref: &Collector{
					opts: &option{},
				},
			},
			tags:                 map[string]string{},
			expectFeedMetricName: "statsd/-/-",
		},

		{
			name: "no_tags",
			acc: &accumulator{
				ref: &Collector{
					opts: &option{
						statsdSourceKey: "source_key",
						statsdHostKey:   "host_key",
					},
				},
			},
			tags:                 map[string]string{},
			expectFeedMetricName: "statsd/-/-",
		},

		{
			name: "default_config_report",
			acc: &accumulator{
				ref: &Collector{
					opts: &option{},
				},
			},
			tags: map[string]string{
				"source_key": "tomcat",
				"host_key":   "cn-shanghai-sq5ei",
			},
			expectFeedMetricName: "statsd/-/-",
		},

		{
			name: "no_source_key",
			acc: &accumulator{
				ref: &Collector{
					opts: &option{
						statsdSourceKey: "source_key",
						statsdHostKey:   "host_key",
					},
				},
			},
			tags: map[string]string{
				"host_key": "cn-shanghai-sq5ei",
			},
			expectFeedMetricName: "statsd/-/cn-shanghai-sq5ei",
		},

		{
			name: "no_host_key",
			acc: &accumulator{
				ref: &Collector{
					opts: &option{
						statsdSourceKey: "source_key",
					},
				},
			},
			tags: map[string]string{
				"source_key": "tomcat",
			},
			expectFeedMetricName: "statsd/tomcat/-",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			acc := tc.acc
			acc.doFeedMetricName(tc.tags)
			require.Equal(t, tc.expectFeedMetricName, acc.feedMetricName)
		})
	}
}
