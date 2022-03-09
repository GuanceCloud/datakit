package pythond

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
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
	originInstallDir := datakit.InstallDir
	originGitReposDir := datakit.GitReposDir
	originPythonDDir := datakit.PythonDDir
	originPythonCoreDir := datakit.PythonCoreDir

	datakit.InstallDir = "/usr/local/datakit"
	datakit.GitReposDir = filepath.Join(datakit.InstallDir, datakit.StrGitRepos)
	datakit.PythonDDir = filepath.Join(datakit.InstallDir, datakit.StrPythonD)
	datakit.PythonCoreDir = filepath.Join(datakit.PythonDDir, datakit.StrPythonCore)

	scriptRoot := `['/usr/local/datakit/gitrepos/conf/python.d/framework']`
	scriptName := "mytest"

	cli := getCliPyScript(scriptRoot, scriptName)

	expectMD5 := "beb828f059208df3647fb0d068eca8b8"

	fmt.Println(cli)
	assert.Equal(t, expectMD5, md5sum(cli), "md5 not equal!")

	datakit.InstallDir = originInstallDir
	datakit.GitReposDir = originGitReposDir
	datakit.PythonDDir = originPythonDDir
	datakit.PythonCoreDir = originPythonCoreDir
}

//------------------------------------------------------------------------------

// go test -v -timeout 30s -run ^TestGetFilteredPyModules$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/pythond
func TestGetFilteredPyModules(t *testing.T) {
	originInstallDir := datakit.InstallDir
	originGitReposDir := datakit.GitReposDir

	datakit.InstallDir = "/usr/local/datakit"
	datakit.GitReposDir = filepath.Join(datakit.InstallDir, datakit.StrGitRepos)

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

	datakit.InstallDir = originInstallDir
	datakit.GitReposDir = originGitReposDir
}

//------------------------------------------------------------------------------

// go test -v -timeout 30s -run ^TestSearchPythondDir$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/pythond
func TestSearchPythondDir(t *testing.T) {
	originInstallDir := datakit.InstallDir
	originGitReposDir := datakit.GitReposDir
	originPythonDDir := datakit.PythonDDir

	datakit.InstallDir = "/usr/local/datakit"
	datakit.GitReposDir = filepath.Join(datakit.InstallDir, datakit.StrGitRepos)
	datakit.PythonDDir = filepath.Join(datakit.InstallDir, datakit.StrPythonD)

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
			dir := searchPythondDir(tc.pythonModule, tc.enabledRepos, &pythondMockerTest{})

			assert.Equal(t, tc.expect, dir)
		})
	}

	datakit.InstallDir = originInstallDir
	datakit.GitReposDir = originGitReposDir
	datakit.PythonDDir = originPythonDDir
}

//------------------------------------------------------------------------------

var (
	dataIsDir, dataGitHasEnabled bool
	dataExistFiles               map[string]struct{}
	dataFolderList               []string
)

func resetVars() {
	dataIsDir = false
	dataGitHasEnabled = false

	dataExistFiles = map[string]struct{}{}

	dataFolderList = []string{}
}

type pythondMockerTest struct{}

func (*pythondMockerTest) IsDir(ph string) bool {
	return dataIsDir
}

func (*pythondMockerTest) FileExist(ph string) bool {
	_, ok := dataExistFiles[ph]
	return ok
}

func (*pythondMockerTest) GetFolderList(root string, deep int) (folders, files []string, err error) {
	return nil, dataFolderList, nil
}

func (*pythondMockerTest) GitHasEnabled() bool {
	return dataGitHasEnabled
}

//------------------------------------------------------------------------------

// go test -v -timeout 30s -run ^TestGetScriptNameRoot$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/pythond
func TestGetScriptNameRoot(t *testing.T) {
	originInstallDir := datakit.InstallDir
	originGitReposDir := datakit.GitReposDir

	datakit.InstallDir = "/usr/local/datakit"
	datakit.GitReposDir = filepath.Join(datakit.InstallDir, datakit.StrGitRepos)

	cases := []struct {
		name          string
		configRepos   []*config.GitRepository
		gitHasEnabled bool
		isDir         bool
		dirs          []string
		existFiles    map[string]struct{}
		folderList    []string
		expect        map[string]string
	}{
		{
			name: "get_script_normal",
			configRepos: []*config.GitRepository{
				{
					Enable: true,
					URL:    "ssh://git@github.com:9000/path/to/repository.git",
				},
			},
			gitHasEnabled: true,
			isDir:         true,
			dirs:          []string{"framework"},
			existFiles:    map[string]struct{}{},
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
			resetVars()
			config.Cfg.GitRepos.Repos = tc.configRepos
			dataGitHasEnabled = tc.gitHasEnabled
			dataIsDir = tc.isDir
			dataExistFiles = tc.existFiles
			dataFolderList = tc.folderList

			if len(tc.configRepos) > 0 {
				config.InitGitreposDir()
			}

			scriptName, scriptRoot, err := getScriptNameRoot(tc.dirs, &pythondMockerTest{})

			assert.NoError(t, err, "getScriptNameRoot error")
			mVal := map[string]string{
				"scriptName": scriptName,
				"scriptRoot": scriptRoot,
			}
			assert.Equal(t, tc.expect, mVal, "getScriptNameRoot not equal")
		})
	}

	datakit.InstallDir = originInstallDir
	datakit.GitReposDir = originGitReposDir
}

//------------------------------------------------------------------------------
