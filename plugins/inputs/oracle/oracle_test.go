package oracle

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func TestRun(t *testing.T) {
	arr, err := config.LoadInputConfigFile("./oracle.conf", func() inputs.Input {
		return &Input{}
	})
	if err != nil {
		t.Fatalf("%s", err)
	}

	o, ok := arr[0].(*Input)
	if !ok {
		t.Error("expect *Input")
	}

	t.Log("args ====>", o.Args)
	o.Run()
}
