package main

import (
	"fmt"
	"testing"
)

func TestCfg(t *testing.T) {

	data := `
{
	"log":       "/usr/local/cloudcare/DataFlux/datakit/oraclemonitor.log",
  "log_level": "info",
	"oracles" : [
		{
			"interval" : "10s",
			"metricName" : "",
			"instanceId" : "",
			"instanceDesc" : "",
			"host" : "",
			"port" : "",
			"username" : "",
			"password" : "",
			"server" : "",
			"type": "singleInstance"
		}
	]
}`

	cfg := loadCfg([]byte(data))
	fmt.Printf("cfg: %+#v\n", cfg)
}
