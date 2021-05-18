package k8s

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/prom"
)

func TestMain(t *testing.T) {
	io.TestOutput()

	p := prom.Prom{
		URL:            "http://172.16.2.42:32136/metrics",
		Interval:       "10s",
		Tags:           map[string]string{"TestTags": "TestValue"},
		InputName:      inputName,
		CatalogStr:     inputName,
		SampleCfg:      sampleCfg,
		IgnoreFunc:     ignore,
		PromToNameFunc: nil,
	}

	p.Run()
}
