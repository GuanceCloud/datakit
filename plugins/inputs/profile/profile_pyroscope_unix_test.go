// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows && !arm && !386
// +build !windows,!arm,!386

package profile

import (
	"os"
	"testing"

	"github.com/GuanceCloud/cliutils"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
)

// 检查是不是开发机，如果不是开发机，则直接退出。开发机上需要定义 LOCAL_UNIT_TEST 环境变量。
func checkDevHost() bool {
	if envs := os.Getenv("LOCAL_UNIT_TEST"); envs == "" {
		return false
	}
	return true
}

// go test -v -timeout 30s -run ^TestPyroscopeRun$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/profile
func TestPyroscopeRun(t *testing.T) {
	if !checkDevHost() {
		return
	}

	pyro := pyroscopeOpts{
		URL: "0.0.0.0:4040",
	}
	config.Cfg.DataWayCfg = &dataway.DataWayCfg{URLs: []string{"http://<GATEWAY>?token=<TOKEN>"}}
	err := config.Cfg.SetupDataway()
	if err != nil {
		panic(err)
	}
	ipt := &Input{
		semStop: cliutils.NewSem(),
	}
	err = pyro.run(ipt)
	if err != nil {
		panic(err)
	}
}

// go test -v -timeout 30s -run ^Test_getLangFamilyFromSpyName$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/profile
func Test_getLangFamilyFromSpyName(t *testing.T) {
	cases := []struct {
		name    string
		spyName string
		expect  string
	}{
		{
			name:    "ebpf",
			spyName: "ebpfspy",
			expect:  "ebpf",
		},
		{
			name:    "go",
			spyName: "goSpy",
			expect:  "go",
		},
		{
			name:    "length_3",
			spyName: "une",
			expect:  "une",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := getLangFamilyFromSpyName(tc.spyName)
			assert.Equal(t, tc.expect, out)
		})
	}
}
