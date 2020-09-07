package confluence

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/prom"
)

func TestMain(t *testing.T) {
	c := Confluence{
		URL:      "http://127.0.0.1:8090/plugins/servlet/prometheus/metrics",
		Interval: "5s",
		Tags:     map[string]string{"TestTags": "TestValue"},
	}

	prom.TestAssert = true
	c.Run()
}
