// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mytargz

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 检查是不是开发机，如果不是开发机，则直接退出。开发机上需要定义 LOCAL_UNIT_TEST 环境变量。
func checkDevHost() bool {
	if envs := os.Getenv("LOCAL_UNIT_TEST"); envs == "" {
		return false
	}
	return true
}

//------------------------------------------------------------------------------

// go test -v -timeout 30s -run ^TestWriteTarFromMap$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/mytargz
func TestWriteTarFromMap(t *testing.T) {
	if !checkDevHost() {
		return
	}

	data := map[string]string{
		"123.p": "4567",
		"abc.p": "defg",
		"中文.p":  "国家",
	}
	dest := "/Users/mac/Downloads/tmp/targz/content.tar.gz"

	err := WriteTarFromMap(data, dest)
	assert.NoError(t, err)
}

// go test -v -timeout 30s -run ^TestReadTarToMap$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/mytargz
func TestReadTarToMap(t *testing.T) {
	if !checkDevHost() {
		return
	}

	src := "/Users/mac/Downloads/tmp/targz/content.tar.gz"

	out, err := ReadTarToMap(src)
	assert.NoError(t, err)

	data := map[string]string{
		"123.p": "4567",
		"abc.p": "defg",
		"中文.p":  "国家",
	}
	assert.Equal(t, data, out)
}
