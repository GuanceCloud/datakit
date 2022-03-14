package config

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func eq(a, b inputs.Input) bool {
	switch x := a.(type) {
	case *cpu:
		return x.eq(b.(*cpu))
	case *disk:
		return x.eq(b.(*disk))
	}
	return false
}

type cpu struct {
	Interval string `toml:"interval"`
	Percpu   bool   `toml:"percpu"`
}

func (*cpu) Catalog() string      { return "" }
func (*cpu) Run()                 {}
func (*cpu) SampleConfig() string { return "no sample" }
func (a *cpu) eq(b *cpu) bool     { return (a.Interval == b.Interval && a.Percpu == b.Percpu) }

type disk struct {
	Interval string `toml:"interval"`
	Percpu   bool   `toml:"percpu"`
}

func (*disk) Catalog() string      { return "" }
func (*disk) Run()                 {}
func (*disk) SampleConfig() string { return "no sample" }
func (a *disk) eq(b *disk) bool    { return (a.Interval == b.Interval && a.Percpu == b.Percpu) }

func TestDoLoadConf(t *testing.T) {
	var _ inputs.Input = &cpu{}

	cases := []struct {
		name         string
		conf         string
		expectInputs map[string][]inputs.Input
	}{

		{
			name: "empty-cpu-conf",
			conf: `[[inputs.cpu]]`,
			expectInputs: map[string][]inputs.Input{
				"cpu": {&cpu{}},
			},
		},

		{
			name: "another-cpu-conf",
			conf: `[inputs]
			  [cpu]
				 a = 10
				[disk]
				 b = 10

				[unknown]
				 c = 10
			`,
		},

		{
			name: "normal-cpu-arr",
			conf: `[[inputs.cpu]]
interval = "10s"
percpu = false
[[inputs.cpu]]
interval = "20s"
percpu = true
`,
			expectInputs: map[string][]inputs.Input{
				"cpu": {
					&cpu{Interval: "10s", Percpu: false},
					&cpu{Interval: "20s", Percpu: true},
				},
			},
		},

		{
			name: "not-input-conf-1",
			conf: `
inputs = "10s" # 故意使用 inputs 试试
percpu = false`,
		},

		{
			name: "not-input-conf-2",
			conf: `
[[inputs]]
	disk = 10
	cpu = "10s"

[[inputs]]
	cpu = ["10s", "20"]
	disk = 10`,
		},

		{
			name: "another-input-conf-format",
			conf: `
[inputs]
	[inputs.cpu]
interval = "15s"
percpu = true
	`,

			expectInputs: map[string][]inputs.Input{
				"cpu": {&cpu{Interval: "15s", Percpu: true}},
			},
		},

		{
			name: "normal-single-cpu",
			conf: `[inputs.cpu]
interval = "11s"
percpu = false`,
			expectInputs: map[string][]inputs.Input{
				"cpu": {&cpu{Interval: "11s", Percpu: false}},
			},
		},

		{
			name: "unknown-input",
			conf: `[inputs.xxx]
interval = "10s"
percpu = false`,
		},

		{
			name: "bad-toml",
			conf: `[inputs.] # bad
interval = "10s"
percpu = false`,
		},

		{
			name: "bad-toml-2",
			conf: `[inputs.cpu]
interval = 10s # bad
percpu = false`,
		},

		{
			name: "bad-unmarshal-type",
			conf: `[inputs.cpu]
interval = 10 # should be string
percpu = false`,
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
			expectInputs: map[string][]inputs.Input{
				"cpu": {
					&cpu{Interval: "10s", Percpu: false},
				},
				"disk": {
					&disk{Interval: "10s", Percpu: false},
				},
			},
		},

		{
			name: "mixed-multiple-inputs",
			conf: `
[[inputs.cpu]]
interval = "10s"
percpu = false

[[inputs.disk]]
interval = "11s"
percpu = false

[[inputs.cpu]]
interval = "10s"
percpu = false

[[inputs.disk]]
interval = "10s"
percpu = false
`,
			expectInputs: map[string][]inputs.Input{
				"cpu": {
					&cpu{Interval: "10s", Percpu: false},
					&cpu{Interval: "10s", Percpu: false},
				},
				"disk": {
					&disk{Interval: "11s", Percpu: false},
					&disk{Interval: "10s", Percpu: false},
				},
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
			}

			tu.Assert(t, len(tc.expectInputs) == len(ret), "expect %d inputs, got %d", len(tc.expectInputs), len(ret))

			for k, arr := range tc.expectInputs {
				for idx, i := range arr {
					tu.Assert(t, eq(i, ret[k][idx]), "not equal: %v <> %v", i, ret[k][idx])
				}
			}

			for k, v := range ret {
				for _, _v := range v {
					t.Logf("%s: %+#v", k, _v)
					t.Logf("sample: %s", _v.SampleConfig())
				}
			}
		})
	}
}
