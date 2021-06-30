package coredns

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/prom"
)

func TestMain(t *testing.T) {
	io.TestOutput()

	p := prom.Prom{
		URL:            "http://127.0.0.1:9901/stats/prometheus",
		Interval:       "10s",
		Tags:           map[string]string{"TestTags": "TestValue"},
		InputName:      inputName,
		CatalogStr:     inputName,
		SampleCfg:      "network",
		IgnoreFunc:     ignore,
		PromToNameFunc: nil,
	}

	p.Run()
}
