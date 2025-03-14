// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package statsd

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseName(t *testing.T) {
	cases := []struct {
		in, out, sep string
	}{
		{
			in:  `jvm.non_heap_memory_max`,
			sep: "_",
			out: "jvm_non_heap_memory_max",
		},
		{
			in:  `jvm.cpu_load.process`,
			sep: "_",
			out: "jvm_cpu_load_process",
		},
		{
			in:  `jvm.buffer_pool.direct.capacity`,
			sep: "_",
			out: "jvm_buffer_pool_direct_capacity",
		},
		{
			in:  `us.west.cpu.load`,
			sep: "_",
			out: "us_west_cpu_load",
		},
	}
	opt := option{}
	s := &Collector{opts: &opt}
	s.Templates = []string{}

	for _, tc := range cases {
		s.opts.metricSeparator = tc.sep
		name, fields, tags := s.parseName(tc.in)
		t.Logf("%s => name: %s, fields: %+#v, tags: %+#v", tc.in, name, fields, tags)
	}
}

func TestParseStatsdLine(t *testing.T) {
	cases := []struct {
		name     string
		line     string
		ddExt    bool
		wantErr  bool
		wantName string
		wantTags map[string]string
	}{
		{
			name:     "basic counter",
			line:     "test.counter:1|c",
			wantName: "test_counter",
			wantTags: map[string]string{"metric_type": "counter"},
		},
		{
			name:     "gauge with sample rate",
			line:     "test.gauge:10.5|g|@0.1",
			wantName: "test_gauge",
			wantTags: map[string]string{"metric_type": "gauge"},
		},
		{
			name:     "set metric",
			line:     "test.set:abc|s",
			wantName: "test_set",
			wantTags: map[string]string{"metric_type": "set"},
		},
		{
			name:     "timing metric",
			line:     "test.timing:350|ms",
			wantName: "test_timing",
			wantTags: map[string]string{"metric_type": "timing"},
		},
		{
			name:     "histogram metric",
			line:     "test.hist:42|h",
			wantName: "test_hist",
			wantTags: map[string]string{"metric_type": "histogram"},
		},
		{
			name:     "distribution metric",
			line:     "test.dist:4.5|d",
			wantName: "test_dist",
			wantTags: map[string]string{"metric_type": "distribution"},
		},
		{
			name:     "with datadog tags",
			line:     "test.counter:1|c|#host:test,env:prod",
			ddExt:    true,
			wantName: "test_counter",
			wantTags: map[string]string{
				"metric_type": "counter",
				"host": "test",
				"env": "prod",
			},
		},
		{
			name:    "invalid format",
			line:    "test.invalid",
			wantErr: true,
		},
		{
			name:    "invalid metric type",
			line:    "test:1|x",
			wantErr: true,
		},
		{
			name:    "invalid value",
			line:    "test:abc|c",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opt := option{
				dataDogExtensions: tc.ddExt,
				metricSeparator: "_",
			}
			col := &Collector{opts: &opt}
			col.Templates = []string{}

			err := col.parseStatsdLine(tc.line)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestParseKeyValue(t *testing.T) {
	cases := []struct {
		input    string
		wantKey  string
		wantVal  string
	}{
		{
			input:   "key=value",
			wantKey: "key",
			wantVal: "value",
		},
		{
			input:   "key=",
			wantKey: "key",
			wantVal: "",
		},
		{
			input:   "=value",
			wantKey: "",
			wantVal: "value",
		},
		{
			input:   "noequals",
			wantKey: "",
			wantVal: "noequals",
		},
		{
			input:   "multiple=equals=signs",
			wantKey: "multiple",
			wantVal: "equals=signs",
		},
		{
			input:   "",
			wantKey: "",
			wantVal: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			key, val := parseKeyValue(tc.input)
			assert.Equal(t, tc.wantKey, key)
			assert.Equal(t, tc.wantVal, val)
		})
	}
}

func TestParser(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		ddExt    bool
		wantName string
		wantTags map[string]string
	}{
		{
			name:     "single metric",
			input:    "test.counter:1|c\n",
			wantName: "test_counter",
			wantTags: map[string]string{"metric_type": "counter"},
		},
		{
			name:     "multiple metrics",
			input:    "test.counter:1|c\ntest.gauge:10|g\n",
			wantName: "test_counter",
			wantTags: map[string]string{"metric_type": "counter"},
		},
		{
			name:     "with empty lines",
			input:    "\ntest.counter:1|c\n\n",
			wantName: "test_counter",
			wantTags: map[string]string{"metric_type": "counter"},
		},
		{
			name:     "datadog event",
			input:    "_e{title,text}:alert\n",
			ddExt:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opt := option{
				dataDogExtensions: tc.ddExt,
				metricSeparator: "_",
			}
			col := &Collector{
				opts: &opt,
				in:   make(chan *Packet, 100),
				done: make(chan struct{}),
			}
			col.Templates = []string{}

			go col.parser(0)

			buf := bytes.NewBufferString(tc.input)
			col.in <- &Packet{
				Buffer: buf,
				Time:   time.Now(),
			}

			time.Sleep(100 * time.Millisecond)
			close(col.done)
		})
	}
}
