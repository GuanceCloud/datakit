package cshark

import (
	"testing"
	"fmt"
	"time"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var msg = `
{
    "device": ["lo"],
    "stream": {
        "duration": "10s",
        "protocol": "http",
        "filter": "'tcp dst port 8080'"
    }
}
`
func TestRun(t *testing.T) {
	t.Run("case-push-data", func(t *testing.T) {
		datakit.InstallDir = "."
		datakit.DataDir = "."
		datakit.OutputFile = "metrics.txt"
		datakit.Exit = cliutils.NewSem()

		s := &Shark{}

		go s.Run()

		for i := 0; i < 10; i++ {
			if err := s.SendOpt(msg); err != nil {
				fmt.Println("err", err)
			}
			time.Sleep(5*time.Second)
		}

		t.Log("ok")
	})
}