package confluence

import (
	"testing"
)

func TestMain(t *testing.T) {
	testAssert = true

	var co = Confluence{
		URL:      "http://127.0.0.1:8090/plugins/servlet/prometheus/metrics",
		Interval: "10s",
	}

	data, err := co.getMetrics()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("data: \n%s\n", data)

}
