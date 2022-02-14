// Package path wrap basic path functions.
package path

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

// go test -v -timeout 30s -run ^TestGetFolderList$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/path
func TestGetFolderList(t *testing.T) {
	if !checkDevHost() {
		return
	}

	const dirr = "/Users/mac/Downloads/project/ent/src/mytest/dir2"
	/*
		.
		├── 1.dir
		│   ├── 2.dir
		│   │   ├── 3.dir
		│   │   └── 3.txt
		│   └── 2.txt
		└── 1.txt

		mkdir -p 1.dir/2.dir/3.dir
		touch 1.txt 1.dir/2.txt 1.dir/2.dir/3.txt
	*/
	expectFolders := []string{
		"/Users/mac/Downloads/project/ent/src/mytest/dir2/1.dir",
	}

	expectFiles := []string{
		"/Users/mac/Downloads/project/ent/src/mytest/dir2/1.dir/2.txt",
		"/Users/mac/Downloads/project/ent/src/mytest/dir2/1.txt",
	}

	folders, files, err := GetFolderList(dirr, 2)
	assert.NoError(t, err, "GetFolderList error!")
	assert.Equal(t, expectFolders, folders, "folders not equal!")
	assert.Equal(t, expectFiles, files, "files not equal!")
}

//------------------------------------------------------------------------------
