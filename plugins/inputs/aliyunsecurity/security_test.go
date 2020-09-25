package aliyunsecurity

import (
	"log"
	"testing"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/sas"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func apiClient() *sas.Client {
	cli, err := sas.NewClientWithAccessKey("cn-hangzhou", "LTAIqo2UBnC4q78J", "t43b4XdKq9Bv50pzSy1yIYiIlwTtvd")
	if err != nil {
		log.Fatalf("create client failed, %s", err)
	}

	return cli
}

//查询对应产品可以获取哪些监控项
func TestDescribeRiskCheckSummary(t *testing.T) {
	security := Security{}

	security.client = apiClient()

	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			// handle
			security.describeRiskCheckSummary("cn-hangzhou")
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		}
	}

	t.Log("ok")
}
