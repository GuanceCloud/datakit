package config

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type cpu struct {
	Interval string `toml:"interval"`
	Percpu   bool   `toml:percpu`
}

func (*cpu) Catalog() string      { return "" }
func (*cpu) Run()                 {}
func (*cpu) SampleConfig() string { return "no sample" }

type disk struct {
	Interval string `toml:"interval"`
	Percpu   bool   `toml:percpu`
}

func (*disk) Catalog() string      { return "" }
func (*disk) Run()                 {}
func (*disk) SampleConfig() string { return "no sample" }

func TestDoLoadConf(t *testing.T) {

	var _ inputs.Input = &cpu{}

	cases := []struct {
		name           string
		conf           string
		expectInputCnt map[string]int
	}{
		{
			name: "normal-cpu-arr",
			conf: `[[inputs.cpu]]
interval = "10s"
percpu = false
[[inputs.cpu]]
interval = "20s"
percpu = true
`,
			expectInputCnt: map[string]int{
				"cpu": 2,
			},
		},

		{
			name: "normal-single-cpu",
			conf: `[inputs.cpu]
interval = "10s"
percpu = false`,
			expectInputCnt: map[string]int{
				"cpu": 1,
			},
		},

		{
			name: "unknown-input",
			conf: `[inputs.xxx]
interval = "10s"
percpu = false`,
			expectInputCnt: map[string]int{
				"xxx": 0,
			},
		},

		{
			name: "bad-toml",
			conf: `[inputs.] # bad
interval = "10s"
percpu = false`,
			expectInputCnt: map[string]int{
				"xxx": 0,
			},
		},

		{
			name: "bad-toml-2",
			conf: `[inputs.cpu]
interval = 10s # bad
percpu = false`,
			expectInputCnt: map[string]int{
				"xxx": 0,
			},
		},

		{
			name: "mixed-inputs",
			conf: `[inputs.cpu]
interval = "10s"
percpu = false

[inputs.disk]
interval = "10s"
percpu = false
`,
			expectInputCnt: map[string]int{
				"cpu":  1,
				"disk": 1,
			},
		},
	}

	creators := map[string]inputs.Creator{
		"cpu": func() inputs.Input {
			return &cpu{}
		},

		"disk": func() inputs.Input {
			return &disk{}
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ret, err := doLoadConf(tc.conf, creators)
			if err != nil {
				t.Logf("doLoadConf: %s", err)
				return
			}

			for k, v := range ret {
				tu.Assert(t, len(v) == tc.expectInputCnt[k],
					"expect got %d %s input, but got %d", tc.expectInputCnt[k], k, len(v))

				for _, _v := range v {
					t.Logf("%s: %+#v", k, _v)
					t.Logf("sample: %s", _v.SampleConfig())
				}
			}
		})
	}
}
