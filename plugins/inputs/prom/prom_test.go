package prom

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

func TestMain(t *testing.T) {
	p := Prom{
		URL:                "http://127.0.0.1:2379/metrics",
		Interval:           "10s",
		Tags:               map[string]string{"TestTags": "TestValue"},
		InputName:          "testing",
		DefaultMeasurement: "testing",
		IgnoreMeasurement:  []string{"testing_grpc_server"},
		log:                logger.DefaultSLogger("testing"),
	}

	p.loadcfg()

	data, err := p.getMetrics()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", data)
}
