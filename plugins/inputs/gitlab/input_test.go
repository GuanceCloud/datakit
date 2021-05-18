package gitlab

import (
	"testing"
)

func TestMain(t *testing.T) {
	instance := newInput()
	instance.URL = "http://127.0.0.1:2080/-/metrics"

	pts, err := instance.gatherMetrics()
	if err != nil {
		t.Fatal(err)
	}

	for _, pt := range pts {
		t.Logf("%s\n", pt.String())
	}
}
