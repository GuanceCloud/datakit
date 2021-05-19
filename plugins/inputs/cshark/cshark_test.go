package cshark

import (
	"fmt"
	"testing"
	"time"
	// "gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var msg = `
{
    "stream": {
        "protocol":"tcp",
        "port": ["3306"],
        "srcIP": ["127.0.0.1"],
        "dstIP": ["127.0.0.1"]
    }
}
`

func TestRun(t *testing.T) {
	t.Run("case-push-data", func(t *testing.T) {
		datakit.InstallDir = "."
		datakit.DataDir = "."
		datakit.OutputFile = "metrics.txt"

		s := &Shark{}
		s.Interval = "3s"
		s.TsharkPath = "/usr/local/bin/tshark"

		go s.Run()

		time.Sleep(time.Second * 10)

		// go func() {
		// 	if err := SendCmdOpt(msg); err != nil {
		// 	   fmt.Println("err", err)
		// 	}
		// }()

		// time.Sleep(10*time.Second)

		if err := SendCmdOpt(msg); err != nil {
			fmt.Println("err", err)
		}

		time.Sleep(100 * time.Second)

		t.Log("ok")
	})
}
