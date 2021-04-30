package kubernetes

import (
	"testing"
)

func TestMain(t *testing.T) {
	var k = newInput()

	k.KubeAPIServerURL = "http://127.0.0.1:8080/metrics"

	pt, err := k.gatherMetrics()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf(pt.String())
}
