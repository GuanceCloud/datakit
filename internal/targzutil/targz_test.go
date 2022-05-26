// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package targzutil

import (
	"os"
	"path/filepath"
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

// go test -v -timeout 30s -run ^TestWriteTarFromMap$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/targzutil
func TestWriteTarFromMap(t *testing.T) {
	if !checkDevHost() {
		return
	}

	cases := []struct {
		name string
		data map[string]string
		dest string
	}{
		{
			name: "normal",
			data: map[string]string{
				"123.p": "4567",
				"abc.p": "defg",
				"中文.p":  "国家",
			},
			dest: "/Users/mac/Downloads/tmp/targz/content.tar.gz",
		},
		{
			name: "dir",
			data: map[string]string{
				filepath.Join("dir1", "123.p"): "4567",
				filepath.Join("dir2", "abc.p"): "defg",
				filepath.Join("dir3", "中文.p"):  "国家",
			},
			dest: "/Users/mac/Downloads/tmp/targz/content_dir.tar.gz",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := WriteTarFromMap(tc.data, tc.dest)
			assert.NoError(t, err)
		})
	}
}

// go test -v -timeout 30s -run ^TestReadTarToMap$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/targzutil
func TestReadTarToMap(t *testing.T) {
	if !checkDevHost() {
		return
	}

	cases := []struct {
		name string
		src  string
		data map[string]string
	}{
		{
			name: "normal",
			src:  "/Users/mac/Downloads/tmp/targz/content.tar.gz",
			data: map[string]string{
				"123.p": "4567",
				"abc.p": "defg",
				"中文.p":  "国家",
			},
		},
		{
			name: "dir",
			src:  "/Users/mac/Downloads/tmp/targz/content_dir.tar.gz",
			data: map[string]string{
				"dir1/123.p": "4567",
				"dir2/abc.p": "defg",
				"dir3/中文.p":  "国家",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := ReadTarToMap(tc.src)
			assert.NoError(t, err)
			assert.Equal(t, tc.data, out)
		})
	}
}
