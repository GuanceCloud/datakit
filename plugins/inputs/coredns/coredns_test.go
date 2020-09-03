package coredns

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/prom"
)

func TestMain(t *testing.T) {
	c := Coredns{
		URL:      "http://127.0.0.1:9153/metrics",
		Interval: "10s",
	}
	prom.TestAssert = true
	c.Run()
}
