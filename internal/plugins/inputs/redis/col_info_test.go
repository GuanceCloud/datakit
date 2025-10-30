// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"os"
	"sort"
	"strings"
	T "testing"

	"github.com/stretchr/testify/assert"
)

func Test_infoGetErrorPoints(t *T.T) {
	tests := []struct {
		name string
		want []string
		tags map[string]string
		k    string
		v    string
	}{
		{
			name: "ok 0",
			tags: map[string]string{"foo": "bar"},
			k:    "errorstat_ERR",
			v:    "count=188",
			want: []string{
				"redis_info,error_type=ERR,foo=bar errorstat=188i",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *T.T) {
			ipt := defaultInput()
			ipt.Tags = tt.tags

			inst := newInstance()
			inst.ipt = ipt
			inst.setup()

			got := inst.getErrorPoints(tt.k, tt.v)

			gotStr := []string{}
			for _, v := range got {
				s := v.LineProto()
				s = s[:strings.LastIndex(s, " ")]
				gotStr = append(gotStr, s)
			}
			sort.Strings(gotStr)

			assert.Equal(t, gotStr, tt.want)
		})
	}
}

func Test_infoGetLatencyPoints(t *T.T) {
	tests := []struct {
		name string
		tags map[string]string
		k,
		v string
		want               []string
		latencyPercentiles bool
	}{
		{
			name: "ok 0",
			tags: map[string]string{"foo": "bar"},
			k:    "latency_percentiles_usec_client|list",
			v:    "p50=23.039,p99=70.143,p99.9=70.143",
			want: []string{
				"redis_info,command_type=client|list,foo=bar,quantile=0.5 latency_percentiles_usec=23.039",
				"redis_info,command_type=client|list,foo=bar,quantile=0.99 latency_percentiles_usec=70.143",
				"redis_info,command_type=client|list,foo=bar,quantile=0.999 latency_percentiles_usec=70.143",
			},
			latencyPercentiles: true,
		},

		{
			name:               "not command stats",
			tags:               map[string]string{"foo": "bar"},
			k:                  "latency_percentiles_usec_client|list",
			v:                  "p50=23.039,p99=70.143,p99.9=70.143",
			want:               nil,
			latencyPercentiles: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *T.T) {
			ipt := defaultInput()
			ipt.Tags = tt.tags

			// setup intput
			ipt.EnableLatencyQuantile = tt.latencyPercentiles
			ipt.mergedTags = tt.tags

			inst := newInstance()
			inst.ipt = ipt
			inst.setup()

			got := inst.getLatencyPoints(tt.k, tt.v)

			var gotStr []string
			for _, v := range got {
				s := v.LineProto()
				s = s[:strings.LastIndex(s, " ")]
				gotStr = append(gotStr, s)
			}
			sort.Strings(gotStr)

			assert.Equal(t, tt.want, gotStr)
		})
	}
}

func Test_getLatencyTagField(t *T.T) {
	tests := []struct {
		name string
		v    string
		want map[string]float64
	}{
		{
			name: "ok 0",
			v:    "p50=23.039,p99=70.143,p99.9=70.143",
			want: map[string]float64{"0.5": 23.039, "0.99": 70.143, "0.999": 70.143},
		},
		{
			name: "ok 1",
			v:    "50=23.039,99=70.143,99.9=70.143",
			want: map[string]float64{"0.5": 23.039, "0.99": 70.143, "0.999": 70.143},
		},
		{
			name: "have dome error",
			v:    "X50=23.039,y99=70.143,99.9=70.143",
			want: map[string]float64{"0.999": 70.143},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *T.T) {
			got := getLatencyTagField(tt.v)
			assert.Equal(t, got, tt.want)
		})
	}
}

func Test_getQuantile(t *T.T) {
	tests := []struct {
		name    string
		s       string
		want    string
		wantErr bool
	}{
		{name: "ok 0", s: "50", want: "0.5", wantErr: false},
		{name: "ok 1", s: "99", want: "0.99", wantErr: false},
		{name: "ok 2", s: "99.9", want: "0.999", wantErr: false},
		{name: "ok 3", s: ".0001", want: "0.000001", wantErr: false},
		{name: "ok 4", s: "1000", want: "10", wantErr: false},
		{name: "error max 6 decimal places", s: ".00001", want: "0.0000001", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *T.T) {
			got, err := getQuantile(tt.s)

			if (err != nil) != tt.wantErr {
				t.Errorf("getQuantile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err != nil) && tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("getQuantile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_infoParseInfoData(t *T.T) {
	ipt := defaultInput()

	t.Run(`maxmemory==0`, func(t *T.T) {
		inst := newInstance()
		inst.ipt = ipt
		pts := inst.parseInfoData(`maxmemory:0
total_system_memory:4171165696`)

		assert.Len(t, pts, 1)
		assert.Equal(t, float64(4171165696), pts[0].Get("maxmemory").(float64))
	})

	t.Run(`maxmemory!=0`, func(t *T.T) {
		inst := newInstance()
		inst.ipt = ipt
		pts := inst.parseInfoData(`maxmemory:1234
total_system_memory:4171165696`)

		assert.Len(t, pts, 1)
		assert.Equal(t, float64(1234), pts[0].Get("maxmemory").(float64))
	})

	t.Run(`perc`, func(t *T.T) {
		inst := newInstance()
		inst.ipt = ipt
		pts := inst.parseInfoData(`some_perc:12.3%`)

		assert.Len(t, pts, 1)
		assert.Equal(t, float64(12.3), pts[0].Get("some_perc").(float64), "pt: %s", pts[0].Pretty())
	})

	t.Run(`kbps`, func(t *T.T) {
		inst := newInstance()
		inst.ipt = ipt
		pts := inst.parseInfoData(`some_kbps:12.3`)

		assert.Len(t, pts, 1)
		assert.Equal(t, float64(12300), pts[0].Get("some_kbps").(float64), "pt: %s", pts[0].Pretty())
	})

	t.Run(`tag`, func(t *T.T) {
		inst := newInstance()
		inst.ipt = ipt
		pts := inst.parseInfoData(`os:linux`)

		assert.Len(t, pts, 1)
		assert.Equal(t, "linux", pts[0].Get("os").(string), "pt: %s", pts[0].Pretty())
	})

	t.Run(`drop-keys`, func(t *T.T) {
		inst := newInstance()
		inst.ipt = ipt
		pts := inst.parseInfoData(`process_id:1234`)

		assert.Len(t, pts, 1)
		assert.Nil(t, pts[0].Get("process_id"), "pt: %s", pts[0].Pretty())
	})

	t.Run(`dynamic-fields`, func(t *T.T) {
		inst := newInstance()
		inst.ipt = ipt
		pts := inst.parseInfoData(`field1:1234
field2:123.4
field3:12.3.4
`)

		assert.Len(t, pts, 1)
		assert.Equal(t, float64(1234), pts[0].Get("field1").(float64), "pt: %s", pts[0].Pretty())
		assert.Equal(t, float64(123.4), pts[0].Get("field2").(float64), "pt: %s", pts[0].Pretty())
		assert.Nil(t, pts[0].Get("field3"), "pt: %s", pts[0].Pretty())
	})

	t.Run(`cpu-usage`, func(t *T.T) {
		inst := newInstance()
		inst.ipt = ipt
		// Use server tag as per-node CPU snapshot key
		inst.mergedTags["server"] = "127.0.0.1:6379"
		pts := inst.parseInfoData(`
used_cpu_sys:16.739323
used_cpu_user:21.649967
`)

		assert.Len(t, pts, 1)
		assert.Equal(t, float64(16.739323), pts[0].Get("used_cpu_sys").(float64), "pt: %s", pts[0].Pretty())
		assert.Equal(t, float64(21.649967), pts[0].Get("used_cpu_user").(float64), "pt: %s", pts[0].Pretty())

		// next time info command
		pts = inst.parseInfoData(`
# added 10s to cpu usage
used_cpu_sys:26.739323
used_cpu_user:31.649967
`)

		assert.Len(t, pts, 1)
		assert.Equal(t, float64(26.739323), pts[0].Get("used_cpu_sys").(float64), "pt: %s", pts[0].Pretty())
		assert.Equal(t, float64(31.649967), pts[0].Get("used_cpu_user").(float64), "pt: %s", pts[0].Pretty())

		assert.True(t, pts[0].Get("used_cpu_user_percent").(float64) > 0, "pt: %s", pts[0].Pretty())
		assert.True(t, pts[0].Get("used_cpu_sys_percent").(float64) > 0, "pt: %s", pts[0].Pretty())

		// Per-node CPU snapshot is stored in infoCPULast keyed by server tag
		last, ok := inst.infoCPULast["127.0.0.1:6379"]
		assert.True(t, ok, "cpu snapshot not found for server key")
		if assert.NotNil(t, last) {
			assert.Equal(t, 26.739323, last.sys)
			assert.Equal(t, 31.649967, last.user)
		}

		t.Logf("%s", pts[0].Pretty()) // default interval is 10s, 2nd info added 10's cpu usage, the percent is 100%.
	})

	t.Run(`redis-7.x-master-info-all`, func(t *T.T) {
		inst := newInstance()
		inst.ipt = ipt
		pts := inst.parseInfoData(redisInfos["7.x-master-info-all"])

		for _, pt := range pts {
			t.Logf("%s", pt.Pretty())
		}
	})
}

var redisInfos = map[string]string{
	"7.x": func() string {
		x, err := os.ReadFile("testdata/redis-info-7.x")
		if err != nil {
			panic("should not been here")
		}

		return string(x)
	}(),

	"7.x-master-info-all": func() string {
		x, err := os.ReadFile("testdata/master-info-all-7.x.txt")
		if err != nil {
			panic("should not been here")
		}

		return string(x)
	}(),
}
