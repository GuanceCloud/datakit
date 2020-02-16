package aliyuncms

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/masahide/toml"
	"github.com/siddontang/go-log/log"

	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers/influx"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
)

func TestConfig(t *testing.T) {

	var cfg AliCMS

	data, err := ioutil.ReadFile("./aliyuncms.toml")
	if err != nil {
		log.Fatalf("%s", err)
	}

	err = toml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatalf("%s", err)
	}
}

func TestCredential(t *testing.T) {

	// cmscfg := &CMS{
	// 	RegionID:        `cn-hangzhou`,
	// 	AccessKeyID:     `LTAIu5wzrLOGHdq1`,
	// 	AccessKeySecret: `***`,
	// }

	// ac := &CMSConfig{
	// 	cfg: cmscfg,
	// }

	// if err := ac.initializeAliyunCMS(); err != nil {
	// 	log.Fatalln(err)
	// } else {
	// 	log.Println("check credential ok")
	// }
}

func TestInfluxLine(t *testing.T) {

	m, _ := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
			"age":   1,
			"bv":    true,
			"nn":    "hello",
		},
		time.Unix(0, 0),
	)

	serializer := influx.NewSerializer()
	output, err := serializer.Serialize(m)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("%s", string(output))
}

func TestMetricInfo(t *testing.T) {

	client, err := cms.NewClientWithAccessKey("cn-hangzhou", "LTAIlsWpTrg1vUf4", "dy5lQzWpU17RDNHGCj84LBDhoU9LVU")

	namespace := "acs_ecs_dashboard"
	metricname := "IOPSUsage"

	request := cms.CreateDescribeMetricMetaListRequest()
	request.Scheme = "https"

	request.Namespace = namespace
	request.MetricName = metricname
	request.PageSize = requests.NewInteger(100)

	response, err := client.DescribeMetricMetaList(request)
	if err != nil {
		fmt.Print(err.Error())
	}

	if response.Resources.Resource != nil {
		for _, res := range response.Resources.Resource {
			fmt.Printf("%s: Periods=%s Statistics=%s\n", res.MetricName, res.Periods, res.Statistics)
		}
	}
}

func TestMetricList(t *testing.T) {

	client, err := cms.NewClientWithAccessKey("cn-hangzhou", "LTAIlsWpTrg1vUf4", "dy5lQzWpU17RDNHGCj84LBDhoU9LVU")

	namespace := "acs_ecs_dashboard"

	request := cms.CreateDescribeMetricMetaListRequest()
	request.Scheme = "https"

	request.Namespace = namespace
	request.PageSize = requests.NewInteger(100)

	response, err := client.DescribeMetricMetaList(request)
	if err != nil {
		fmt.Print(err.Error())
	}

	if response.Resources.Resource != nil {
		for _, res := range response.Resources.Resource {
			fmt.Printf("%s: Periods=%s Statistics=%s\n", res.MetricName, res.Periods, res.Statistics)
		}
	}
}

func TestGetMetrics(t *testing.T) {
	client, err := cms.NewClientWithAccessKey("cn-hangzhou", "LTAIlsWpTrg1vUf4", "dy5lQzWpU17RDNHGCj84LBDhoU9LVU")

	request := cms.CreateDescribeMetricListRequest()
	request.Scheme = "https"

	request.MetricName = "CPUUtilization"
	request.Namespace = "acs_ecs_dashboard"
	request.Period = "60"
	request.StartTime = "1574921940000"
	request.EndTime = "1574922000000"
	request.Dimensions = `[{"instanceId": "i-bp15wj5w33t8vfxi7z3d"}]`

	response, err := client.DescribeMetricList(request)
	if err != nil {
		fmt.Print(err.Error())
	}

	var datapoints []map[string]interface{}
	if err = json.Unmarshal([]byte(response.Datapoints), &datapoints); err != nil {
		log.Fatalf("failed to decode response datapoints: %v", err)
	}

	log.Infof("Count: %v", len(datapoints))
	for _, dp := range datapoints {
		log.Infof("instanceId:%v,  Minimum:%v, Average:%v, timestamp:%v", dp["instanceId"], dp["Minimum"], dp["Average"], dp["timestamp"])
	}
}

func TestSvr(t *testing.T) {

	var cfg AliCMS

	data, err := ioutil.ReadFile("./aliyuncms.toml")
	if err != nil {
		log.Fatalf("%s", err)
	}

	err = toml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatalf("%s", err)
	}

	cfg.Start(nil)
}
