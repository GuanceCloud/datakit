package pythond

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
)

// go test -v -timeout 30s -run ^TestPythonD$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/pythond
// func TestPythonD(t *testing.T) {
// 	config.Cfg.GitRepos.Repos = []*config.GitRepository{
// 		{
// 			Enable: true,
// 			URL:    "ssh://git@gitlab.jiagouyun.com:40022/jack/conf.git",
// 		},
// 	}
// 	datakit.PythonDDir = "/Users/mac/Downloads/project/ent/src/gitlab.jiagouyun.com/cloudcare-tools/datakitpy"
// 	pe := &PythonDInput{
// 		Name:    "some-python-inputs",
// 		Cmd:     "python3",
// 		Dirs:    []string{"framework"},
// 		semStop: cliutils.NewSem(),
// 	}
// 	pe.Run()
// }

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
