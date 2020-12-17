package cshark

import (
	"testing"
	"fmt"
	"time"
	// "gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var msg = `
{
    "device": ["lo"],
    "stream": {
        "duration": "10s",
        "protocol": "udp",
        "filter": "'tcp dst port 8080'"
    }
}
`
func TestRun(t *testing.T) {
	t.Run("case-push-data", func(t *testing.T) {
		datakit.InstallDir = "."
		datakit.DataDir = "."
		datakit.OutputFile = "metrics.txt"

		s := &Shark{}

		go s.Run()

		time.Sleep(time.Second*1)

		if err := SendCmdOpt(msg); err != nil {
			fmt.Println("err", err)
		}

		time.Sleep(30*time.Second)

		t.Log("ok")
	})
}