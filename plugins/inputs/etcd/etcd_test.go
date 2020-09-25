package etcd

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/prom"
)

func TestMain(t *testing.T) {
	e := Etcd{
		URL:      "http://127.0.0.1:2379/metrics",
		Interval: "10s",
		Tags:     map[string]string{"TestTags": "TestValue"},
	}
	prom.TestAssert = true
	e.Run()
}
