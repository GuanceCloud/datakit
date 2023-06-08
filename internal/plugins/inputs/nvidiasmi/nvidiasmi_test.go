// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package nvidiasmi

import (
	"fmt"
	"testing"
)

// 测试抓取+转换程序
func TestCollect(t *testing.T) {
	ipt := newDefaultInput()
	err := ipt.Collect()
	fmt.Println("err === ", err)
}

// 测试抓取程序
func TestCombinedOutputTimeout(t *testing.T) {
	ipt := newDefaultInput()
	ret, err := ipt.getBytes("/usr/bin/nvidia-smi")
	fmt.Println("err === ", err)
	fmt.Println("ret === ", ret)
	_ = ret
}
