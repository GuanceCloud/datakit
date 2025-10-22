// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package statsd

import (
	T "testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseName(t *T.T) {
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

func TestParseLine(t *T.T) {
	t.Run(`with-|T`, func(t *T.T) {
		line := "namespace.test_gauge:21|#globalTags,globalTags2,tag1,tag2|T1658997712"
		c, err := NewCollector(nil, nil, WithProtocol("udp"), WithDataDogExtensions(true))
		require.NoError(t, err)
		assert.NoError(t, c.parseStatsdLine(line))
	})

	t.Run(`with-dd-tags`, func(t *T.T) {
		line := "namespace.test_gauge:21|g|#globalTags,globalTags2,tag1,tag2"
		c, err := NewCollector(nil, nil, WithProtocol("udp"), WithDataDogExtensions(true))
		require.NoError(t, err)
		assert.NoError(t, c.parseStatsdLine(line))
	})
}
