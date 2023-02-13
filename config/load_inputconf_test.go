// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tu "github.com/GuanceCloud/cliutils/testutil"
	"github.com/stretchr/testify/assert"
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
			ret, err := LoadSingleConf(tc.conf, creators)
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

// go test -v -timeout 30s -run ^Test_SearchDir$ gitlab.jiagouyun.com/cloudcare-tools/datakit/config
func Test_SearchDir(t *testing.T) {
	cases := []struct {
		name                        string
		testDirName                 string
		preparedDirs, preparedFiles []string
		suffix                      string
		ignoreDirs                  []string
		out                         []string
	}{
		{
			name:        "normal",
			testDirName: "test_data",
			preparedDirs: []string{
				"dir1",
				"dir2/.git",
			},
			preparedFiles: []string{
				"a.conf",
				"dir1/b.conf",
				"dir1/c.conf",
				"dir1/d.conf.sample",
				"dir2/e.conf",
				"dir2/.git/f.conf",
			},
			suffix:     ".conf",
			ignoreDirs: []string{".git"},
			out: []string{
				"/a.conf",
				"/dir1/b.conf",
				"/dir1/c.conf",
				"/dir2/e.conf",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rootDir := prepareDirs(tc.testDirName, tc.preparedDirs, tc.preparedFiles)
			fmt.Printf("rootDir = %s\n", rootDir)

			out := SearchDir(rootDir, tc.suffix, tc.ignoreDirs...)
			for k, v := range out {
				stripped := strings.Replace(v, rootDir, "", 1)
				out[k] = stripped
			}
			assert.Equal(t, tc.out, out)

			if err := os.RemoveAll(rootDir); err != nil {
				panic(err)
			}
		})
	}
}

func prepareDirs(testDirName string, arrDirs, arrFiles []string) string {
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	testDir := filepath.Join(path, testDirName)

	for _, v := range arrDirs {
		path := filepath.Join(testDir, v)
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			panic(err)
		}
	}

	for _, v := range arrFiles {
		path := filepath.Join(testDir, v)
		if err := TouchFile(path); err != nil {
			panic(err)
		}
	}

	return testDir
}

func TouchFile(name string) error {
	file, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	return file.Close()
}
