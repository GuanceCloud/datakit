package envoy

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/prom"
)

func TestMain(t *testing.T) {
	e := Envoy{
		URL:      "http://127.0.0.1:9901/stats/prometheus",
		Interval: "10s",
		Tags:     map[string]string{"TestTags": "TestValue"},
	}

	prom.TestAssert = true
	e.Run()
}
