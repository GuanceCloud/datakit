package prom

import (
	"testing"
)

func TestMain(t *testing.T) {
	p := Prom{
		URL:               "http://127.0.0.1:2379/metrics",
		Interval:          "10s",
		Tags:              map[string]string{"TestTags": "TestValue"},
		InputName:         "testing",
		IgnoreMeasurement: []string{"testing_grpc_server"},
	}

	p.loadcfg()
	data, err := p.getMetrics()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", data)
}
