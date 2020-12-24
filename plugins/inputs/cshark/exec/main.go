package main

import (
	"fmt"
	"time"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/cshark"
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
func main() {
	datakit.DataDir = "."
	datakit.OutputFile = "metrics.txt"
	datakit.Exit = cliutils.NewSem()

	s := &cshark.Shark{}

	s.Run()

	if err := s.SendOpt(msg); err != nil {
		fmt.Println("err", err)
	}

	time.Sleep(10*time.Second)
}