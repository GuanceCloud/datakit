package nginx

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"testing"
	"time"
)

func TestGetStubStatusModule(t *testing.T) {
	n := &Input{
		Url: "http://0.0.0.0:8080/nginx_status",
	}
	client, err := n.createHttpClient()
	if err != nil {
		l.Fatal(err)
	}
	n.client = client

	n.getStubStatusModuleMetric()
}

func TestGetVTSMetric(t *testing.T) {
	n := &Input{
		Url:             "http://10.100.65.53:8888/status/format/json",
		ResponseTimeout: datakit.Duration{Duration: time.Second * 20},
	}
	l = logger.SLogger(inputName)

	client, err := n.createHttpClient()
	if err != nil {
		l.Fatal(err)
	}
	n.client = client
	n.getVTSMetric()
}
