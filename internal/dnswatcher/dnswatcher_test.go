// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package dnswatcher contains dns watcher control logic
package dnswatcher

import (
	"os"
	T "testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//------------------------------------------------------------------------------

// 检查是不是开发机，如果不是开发机，则直接退出。开发机上需要定义 LOCAL_UNIT_TEST 环境变量。
func checkDevHost() bool {
	if envs := os.Getenv("LOCAL_UNIT_TEST"); envs == "" {
		return false
	}
	return true
}

//------------------------------------------------------------------------------

// go test -v -timeout 30s -run ^TestGetCheckInterval$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dnswatcher
func TestGetCheckInterval(t *T.T) {
	cases := []struct {
		name string
		in   string
		out  time.Duration
	}{
		{
			name: "normal",
			in:   "5s",
			out:  5 * time.Second,
		},
		{
			name: "empty",
			in:   "",
			out:  defaultCheckInterval,
		},
		{
			name: "string",
			in:   "string",
			out:  defaultCheckInterval,
		},
		{
			name: "number",
			in:   "123",
			out:  defaultCheckInterval,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *T.T) {
			du := getCheckInterval(tc.in)
			assert.Equal(t, tc.out, du)
		})
	}
}

//------------------------------------------------------------------------------

type BaiduImpl struct {
	t *T.T
}

func (*BaiduImpl) GetDomain() string {
	return "baidu.com"
}

var baiduIPs = []string{"220.181.38.251", "220.181.38.148"}

func (*BaiduImpl) GetIPs() []string {
	return baiduIPs
}

func (x *BaiduImpl) SetIPs([]string) {
	x.t.Log("SetIPs")
}

func (x *BaiduImpl) Update() error {
	x.t.Log("Update")
	return nil
}

// Make sure BaiduImpl implements the IDNSWatcher interface
var _ IDNSWatcher = new(BaiduImpl)

type QQImpl struct {
	t *T.T
}

func (*QQImpl) GetDomain() string {
	return "qq.com"
}

func (*QQImpl) GetIPs() []string {
	return []string{}
}

func (x *QQImpl) SetIPs([]string) {
	x.t.Log("SetIPs")
}

func (x *QQImpl) Update() error {
	x.t.Log("Update")
	return nil
}

// Make sure QQImpl implements the IDNSWatcher interface
var _ IDNSWatcher = new(QQImpl)

// go test -v -timeout 30s -run ^TestCheckDNSChanged$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dnswatcher
func TestCheckDNSChanged(t *T.T) {
	if !checkDevHost() {
		return
	}

	cases := []struct {
		name     string
		in       IDNSWatcher
		outBool  bool
		outArray []string
	}{
		{
			name: "baidu",
			in: &BaiduImpl{
				t: t,
			},
			outBool:  false,
			outArray: baiduIPs,
		},
		{
			name:     "qq",
			in:       &QQImpl{t: t},
			outBool:  true,
			outArray: []string{"183.3.226.35", "123.151.137.18", "61.129.7.47"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *T.T) {
			outBool, outArray := checkDNSChanged(tc.in)
			assert.Equal(t, tc.outBool, outBool)
			assert.ElementsMatch(t, tc.outArray, outArray)
		})
	}
}

//------------------------------------------------------------------------------
