package statsd

import (
	"testing"
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

		//{
		//	in:  `jvm.non_heap_memory_init:7667712`,
		//	out: "",
		//},
	}

	s := defaultInput()
	s.Templates = []string{}

	for _, tc := range cases {

		s.MetricSeparator = tc.sep

		name, fields, tags := s.parseName(tc.in)
		t.Logf("%s => name: %s, fields: %+#v, tags: %+#v", tc.in, name, fields, tags)
	}
}
