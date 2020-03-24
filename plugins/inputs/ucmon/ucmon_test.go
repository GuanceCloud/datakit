package ucmon

import (
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/toml"

	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

func loadConfig() (*ucloud.Config, *auth.Credential) {
	cfg := ucloud.NewConfig()

	credential := auth.NewCredential()
	credential.PrivateKey = "o3kgCI6D55TMPSX5cyvdW6pLfnSELJnaOrJhPwl2HWsJuSr2bhR8pr0PYVgsCZz5"
	credential.PublicKey = "yWNvxpEq8hA42qABkaeRxloYhROUmAnFUgOkfJ0K1hfw36mG6aIqUk7dPts="

	return &cfg, &credential
}

func TestConfig(t *testing.T) {
	var cfg ucMonitorAgent

	cfg.Instances = []*ucInstance{
		&ucInstance{
			PublicKey:  "xxx",
			PrivateKey: "xxx",
			Region:     "xxx",
			Zone:       "xxx",
			ProjectID:  "xxx",
			Resource: []*ucResource{
				&ucResource{
					ResourceID:   "xxx",
					ResourceType: "xxx",
					Metrics: []*ucMetric{
						&ucMetric{
							MetricName: "xxx",
						},
						&ucMetric{
							MetricName: "yyy",
						},
					},
				},
			},
		},
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		log.Fatalf("%s", err)
	}
	log.Printf("cfg: %s", string(data))
}

func TestRegion(t *testing.T) {

	cfg, credential := loadConfig()
	uhostClient := uhost.NewClient(cfg, credential)
	_ = uhostClient

	cli := ucloud.NewClient(cfg, credential)
	req := cli.NewGenericRequest()
	req.SetAction("GetRegion")

	resp, err := cli.GenericInvoke(req)
	if err != nil {
		log.Fatalf("%s", err)
	}
	resContent := resp.GetPayload()
	//log.Printf("type: %s", reflect.TypeOf(resContent["Regions"]))
	if regions, ok := resContent["Regions"].([]interface{}); ok {
		for _, reg := range regions {
			if regmap, ok := reg.(map[string]interface{}); ok {
				log.Printf("%v", regmap[`RegionName`])
			}
		}
	} else {
		log.Printf("%v", regions)
	}
}

func TestProjectList(t *testing.T) {
	cfg, credential := loadConfig()
	cli := ucloud.NewClient(cfg, credential)
	req := cli.NewGenericRequest()
	req.SetAction("GetProjectList")
	m := map[string]interface{}{}
	m["IsFinance"] = "Yes"
	req.SetPayload(m)

	resp, err := cli.GenericInvoke(req)
	if err != nil {
		log.Fatalf("%s", err)
	}
	_ = resp
	log.Printf("resp: %v", resp)
}

func TestMetricMeta(t *testing.T) {
	cfg, credential := loadConfig()
	cli := ucloud.NewClient(cfg, credential)
	req := cli.NewGenericRequest()
	req.SetAction("DescribeResourceMetric")
	m := map[string]interface{}{}
	m["ResourceType"] = "uhost"
	req.SetPayload(m)

	resp, err := cli.GenericInvoke(req)
	if err != nil {
		log.Fatalf("%s", err)
	}

	payload := resp.GetPayload()
	if dataset, ok := payload["DataSet"].([]interface{}); ok {
		for _, md := range dataset {
			if metricItem, ok2 := md.(map[string]interface{}); ok2 {
				log.Printf("%s, type=%s, unit=%s, Frequency=%v", metricItem["MetricName"], metricItem["Type"], metricItem["Unit"], metricItem["Frequency"])
			}
		}
	}

}

func TestGetMetrics(t *testing.T) {
	cfg, credential := loadConfig()
	cli := ucloud.NewClient(cfg, credential)
	req := cli.NewGenericRequest()
	req.SetAction("GetMetric")
	req.SetRegion("cn-sh2")
	req.SetZone("cn-sh2-03")
	req.SetProjectId("org-lzdm2g")
	m := map[string]interface{}{}
	m["ResourceType"] = "uhost"
	m["ResourceId"] = "uhost-500cdl5d"
	m["MetricName"] = "IORead"
	m["BeginTime"] = time.Now().Truncate(time.Minute).Add(-5 * time.Minute).Unix()
	m["EndTime"] = time.Now().Truncate(time.Minute).Unix()
	req.SetPayload(m)

	resp, err := cli.GenericInvoke(req)
	if err != nil {
		t.Errorf("GenericInvoke failed: %s", err)
	}

	payload := resp.GetPayload()
	log.Printf("%s", reflect.TypeOf(payload["DataSets"]))
	if mapData, ok := payload["DataSets"].(map[string]interface{}); ok {
		for k, v := range mapData {
			log.Printf("key=%s", k)
			if vals, ok2 := v.([]interface{}); ok2 {
				for _, mv := range vals {
					if mapv, ok3 := mv.(map[string]interface{}); ok3 {
						log.Printf("%v", mapv)
					}
				}
			}
		}
	}
}
