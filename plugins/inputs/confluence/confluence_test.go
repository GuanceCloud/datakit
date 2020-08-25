package confluence

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

func __init() {
	logger.SetGlobalRootLogger("", logger.DEBUG, logger.OPT_DEFAULT)
	l = logger.SLogger(inputName)
	testAssert = true
}

func TestMain(t *testing.T) {

	__init()

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
