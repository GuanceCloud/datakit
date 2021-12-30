package pythond

import (
	"os"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
)

// 开发的电脑名。把这个改成你的电脑名则可跑全所有 case。
// 根据 case 不同可能需要配置一定量的环境。具体要配多少自己跑跑就知道。怎么配看代码！
const devHostName = "macdeMacBook-Air.local"

// 检查是不是开发机，如果不是开发机，则直接退出。
func checkDevHost() bool {
	name, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	if name != devHostName {
		return false
	}
	return true
}

// go test -v -timeout 30s -run ^TestPythonD$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/pythond
func TestPythonD(t *testing.T) {
	if !checkDevHost() {
		return
	}

	config.Cfg.GitRepos.Repos = []*config.GitRepository{
		{
			Enable: true,
			URL:    "ssh://git@gitlab.jiagouyun.com:40022/jack/conf.git",
		},
	}
	datakit.PythonDDir = "/Users/mac/Downloads/project/ent/src/gitlab.jiagouyun.com/cloudcare-tools/datakitpy"
	pe := &PythonDInput{
		Name:    "some-python-inputs",
		Cmd:     "python3",
		Dirs:    []string{"framework"},
		semStop: cliutils.NewSem(),
	}
	pe.Run()
}

// go test -v -timeout 30s -run ^TestGetPyModules$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/pythond
func TestGetPyModules(t *testing.T) {
	ss := getPyModules("/usr/local/datakit/gitrepos/python.d/framework")
	ss = dkstring.GetUniqueArray(ss)
	t.Logf("pyModules = %v", ss)
}

// go test -v -timeout 30s -run ^TestSearchPythondDir$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/pythond
func TestSearchPythondDir(t *testing.T) {
	dirs := []string{"conf"}
	s := searchPythondDir("framework", dirs)
	t.Logf("searchPythondDir: %s", s)
}
