package statsd

import (
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
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

	acc := &accumulator{}
	s := defaultInput()
	acc.ref = s
	s.acc = acc

	for _, tc := range cases {
		acc.points = acc.points[:0] //clear cache

		s.MetricMapping = tc.mmap
		s.DropTags = tc.dropTags
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
