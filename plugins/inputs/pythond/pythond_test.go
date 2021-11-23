package pythond

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

// go test -v -timeout 30s -run ^TestPythonD$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/pythond
func TestPythonD(t *testing.T) {
	// config.Cfg.GitRepos.Repos[0].
	datakit.PythonDDir = "/Users/mac/Downloads/project/ent/src/gitlab.jiagouyun.com/cloudcare-tools/datakitpy"
	pe := &PythonDInput{
		Name:    "some-python-inputs",
		Cmd:     "python3",
		Dirs:    []string{"framework", "datakit"},
		semStop: cliutils.NewSem(),
	}
	pe.Run()
}
