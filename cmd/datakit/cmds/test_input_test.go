package cmds

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmds/configs"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/all"
)

// go test -v -timeout 30s -run ^TestPrintInputOnce$ gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmds
func TestPrintInputOnce(t *testing.T) {
	configFiles, err := configs.ReleaseFiles()
	if err != nil {
		t.Error(err)
		return
	}

	for _, v := range configFiles {
		t.Log(v)
		if err := printInputOnceCore(t, v); err != nil {
			t.Error(err)
			return
		}
	}

	t.Log("PrintInputOnce test completed!")
}

func printInputOnceCore(t *testing.T, configPath string) error {
	t.Helper()

	FlagCmdLogPath := "/dev/null"

	tryLoadMainCfg()
	setCmdRootLog(FlagCmdLogPath)
	if err := inputDebugger(configPath); err != nil {
		t.Errorf("inputDebugger: %s", err)
		return err
	}

	return nil
}
