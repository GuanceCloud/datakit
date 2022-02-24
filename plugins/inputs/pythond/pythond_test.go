package pythond

import (
	"crypto/md5"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
)

func md5sum(str string) string {
	h := md5.New() //nolint:gosec
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

//------------------------------------------------------------------------------

// go test -v -timeout 30s -run ^TestGetCliPyScript$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/pythond
func TestGetCliPyScript(t *testing.T) {
	scriptRoot := `['/usr/local/datakit/gitrepos/conf/python.d/framework']`
	scriptName := "mytest"

	cli := getCliPyScript(scriptRoot, scriptName)

	expectMD5 := "beb828f059208df3647fb0d068eca8b8"

	assert.Equal(t, expectMD5, md5sum(cli), "md5 not equal!")
}

//------------------------------------------------------------------------------

// go test -v -timeout 30s -run ^TestGetFilteredPyModules$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/pythond
func TestGetFilteredPyModules(t *testing.T) {
	cases := []struct {
		name   string
		files  []string
		root   string
		expect []string
	}{
		{
			name: "standard_and_right",
			files: []string{
				"/usr/local/datakit/gitrepos/repository/python.d/file1.py",
				"/usr/local/datakit/gitrepos/repository/python.d/dir1/file2.py",
			},
			root: "/usr/local/datakit/gitrepos/repository/python.d",
			expect: []string{
				"file1",
				"dir1.file2",
			},
		},

		{
			name: "not_standard_but_right",
			files: []string{
				"/usr/local/datakit/gitrepos/repository/python.d/dir1/dir2/file3.py",
				"/usr/local/datakit/gitrepos/repository/python.d/dir1/dir2/dir3/file4.py",
			},
			root: "/usr/local/datakit/gitrepos/repository/python.d",
			expect: []string{
				"dir2.file3",
				"dir3.file4",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			arr := getFilteredPyModules(tc.files, tc.root)
			assert.Equal(t, tc.expect, arr)
		})
	}
}

//------------------------------------------------------------------------------

var dataIsDir bool

type pathMockerTest struct{}

func (*pathMockerTest) IsDir(ph string) bool {
	return dataIsDir
}

// go test -v -timeout 30s -run ^TestSearchPythondDir$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/pythond
func TestSearchPythondDir(t *testing.T) {
	cases := []struct {
		name         string
		isDIr        bool
		pythonModule string
		enabledRepos []string
		expect       string
	}{
		{
			name:         "not_dir_conf1",
			isDIr:        false,
			pythonModule: "framework",
			enabledRepos: []string{"enabled_conf1"},
			expect:       "/usr/local/datakit/python.d/framework",
		},

		{
			name:         "dir_conf2",
			isDIr:        true,
			pythonModule: "framework",
			enabledRepos: []string{"enabled_conf2"},
			expect:       "/usr/local/datakit/gitrepos/enabled_conf2/python.d/framework",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dataIsDir = tc.isDIr
			dir := searchPythondDir(tc.pythonModule, tc.enabledRepos, &pathMockerTest{})

			assert.Equal(t, tc.expect, dir)
		})
	}
}

//------------------------------------------------------------------------------

var dataExistDirs, dataExistFiles map[string]struct{}

type pathMockExerTest struct{}

func (*pathMockExerTest) IsDir(ph string) bool {
	_, ok := dataExistDirs[ph]
	return ok
}

func (*pathMockExerTest) FileExist(ph string) bool {
	_, ok := dataExistFiles[ph]
	return ok
}

var dataFolderList []string

type folderListMockerTest struct{}

func (*folderListMockerTest) GetFolderList(root string, deep int) (folders, files []string, err error) {
	return nil, dataFolderList, nil
}

// go test -v -timeout 30s -run ^TestGetScriptNameRoot$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/pythond
func TestGetScriptNameRoot(t *testing.T) {
	cases := []struct {
		name        string
		configRepos []*config.GitRepository
		isDIr       bool
		dirs        []string
		existDirs   map[string]struct{}
		existFiles  map[string]struct{}
		folderList  []string
		expect      map[string]string
	}{
		{
			name: "get_script_normal",
			configRepos: []*config.GitRepository{
				{
					Enable: true,
					URL:    "ssh://git@github.com:9000/path/to/repository.git",
				},
			},
			isDIr: true,
			dirs:  []string{"framework"},
			existDirs: map[string]struct{}{
				"/usr/local/datakit/gitrepos/repository/python.d/framework": {},
			},
			existFiles: map[string]struct{}{},
			folderList: []string{
				"/usr/local/datakit/gitrepos/repository/python.d/framework/mytest.py",
			},
			expect: map[string]string{
				"scriptName": "mytest",
				"scriptRoot": "['/usr/local/datakit/gitrepos/repository/python.d/framework']",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			config.Cfg.GitRepos.Repos = tc.configRepos
			dataIsDir = tc.isDIr
			dataExistDirs = tc.existDirs
			dataExistFiles = tc.existFiles
			dataFolderList = tc.folderList

			if len(tc.configRepos) > 0 {
				config.InitGitreposDir()
			}

			scriptName, scriptRoot, err := getScriptNameRoot(tc.dirs, &pathMockerTest{}, &pathMockExerTest{}, &folderListMockerTest{})

			assert.NoError(t, err, "getScriptNameRoot error")
			mVal := map[string]string{
				"scriptName": scriptName,
				"scriptRoot": scriptRoot,
			}
			assert.Equal(t, tc.expect, mVal, "getScriptNameRoot not equal")
		})
	}
}

//------------------------------------------------------------------------------
