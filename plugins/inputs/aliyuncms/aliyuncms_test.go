package aliyuncms

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/influxdata/toml"

	"github.com/influxdata/telegraf/metric"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
)

func apiClient() *cms.Client {
	client, err := cms.NewClientWithAccessKey("cn-hangzhou", "", "")
	if err != nil {
		log.Fatalf("%s", err)
	}
	return client
}

func TestDumpConfig(t *testing.T) {
	var ag CMS

	ag.Project = []*Project{
		&Project{
			Name: "ecs",
			Metrics: &Metric{
				MetricNames: []string{"cpu", "mem"},
				Property: []*Property{
					&Property{
						Name:       "*",
						Period:     60,
						Dimensions: "dddd",
						Tags: map[string]string{
							"key1": "val1",
							"key2": "val2",
						},
					},
					&Property{
						Name:       "p1",
						Period:     60,
						Dimensions: "dddd111",
						Tags: map[string]string{
							"key1": "val1",
						},
					},
				},
			},
		},
		&Project{
			Name: "rds",
			Metrics: &Metric{
				MetricNames: []string{"cpu", "mem"},
				Property: []*Property{
					&Property{
						Name:       "*",
						Period:     60,
						Dimensions: "dddd",
						Tags: map[string]string{
							"key1": "val1",
							"key2": "val2",
						},
					},
					&Property{
						Name:       "p1",
						Period:     60,
						Dimensions: "dddd111",
						Tags: map[string]string{
							"key1": "val1",
						},
					},
				},
			},
		},
	}

	data, err := toml.Marshal(&ag)
	if err != nil {
		t.Error(err)
	}
	log.Printf("%s", string(data))
}

func TestLoadConfig(t *testing.T) {

	cfgData, err := ioutil.ReadFile("test.conf")
	if err != nil {
		t.Error(err)
	}

	var cfg CMS

	err = toml.Unmarshal([]byte(cfgData), &cfg)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		log.Printf("ok")
	}
}

//查询对应产品可以获取哪些监控项
func TestGetMetricMeta(t *testing.T) {

	client, err := cms.NewClientWithAccessKey("cn-hangzhou", "", "")
	if err != nil {
		log.Fatalf("%s", err)
	}

	request := cms.CreateDescribeMetricMetaListRequest()
	request.Scheme = "https"
	request.Namespace = `acs_ecs_dashboard`
	request.MetricName = `CPUUtilization`

	response, err := client.DescribeMetricMetaList(request)
	if err != nil {
		t.Errorf("error: %s", err)
	}

	log.Printf("count=%d", len(response.Resources.Resource))

	if response.Resources.Resource != nil {
		for _, res := range response.Resources.Resource {
			//有些指标的Statistics返回为空
			log.Printf("%s: Periods=%s Statistics=%s, Dimension=%s, Unit=%v", res.MetricName, res.Periods, res.Statistics, res.Dimensions, res.Unit)
		}
	}
}

func TestGetMetricList(t *testing.T) {

	request := cms.CreateDescribeMetricListRequest()
	request.Scheme = "https"
	request.RegionId = "cn-hangzhou"
	request.MetricName = "CPUUtilization"
	request.Namespace = "acs_ecs_dashboard"
	request.Period = "60"
	request.StartTime = "1588156800000"
	request.EndTime = "1588157100000"
	//request.StartTime = "1585209420000"
	//request.EndTime = "1585209600000"

	request.Dimensions = `[{"instanceId":"i-bp1bnmpvz2tuu7odx6cx"}]`
	request.NextToken = ""

	for {
		response, err := apiClient().DescribeMetricList(request)
		if err != nil {
			t.Errorf("DescribeMetricList error, %s", err)
			return
		}

		var datapoints []map[string]interface{}
		if err = json.Unmarshal([]byte(response.Datapoints), &datapoints); err != nil {
			t.Errorf("failed to decode response datapoints: %s", err)
		}

		log.Printf("Datapoints Count = %v", len(datapoints))

		metricName := request.MetricName

		for _, datapoint := range datapoints {

			tags := map[string]string{
				"regionId": request.RegionId,
			}

			fields := make(map[string]interface{})

			if average, ok := datapoint["Average"]; ok {
				fields[formatField(metricName, "Average")] = average
			}
			if minimum, ok := datapoint["Minimum"]; ok {
				fields[formatField(metricName, "Minimum")] = minimum
			}
			if maximum, ok := datapoint["Maximum"]; ok {
				fields[formatField(metricName, "Maximum")] = maximum
			}
			if value, ok := datapoint["Value"]; ok {
				fields[formatField(metricName, "Value")] = value
			}
			if value, ok := datapoint["Sum"]; ok {
				fields[formatField(metricName, "Sum")] = value
			}

			for _, k := range supportedDimensions {
				if kv, ok := datapoint[k]; ok {
					if kvstr, bok := kv.(string); bok {
						tags[k] = kvstr
					} else {
						tags[k] = fmt.Sprintf("%v", kv)
					}
				}
			}

			tm := time.Now()
			switch ft := datapoint["timestamp"].(type) {
			case float64:
				tm = time.Unix((int64(ft))/1000, 0)
			}

			newMetric, _ := metric.New(formatMeasurement(request.Namespace), tags, fields, tm)
			_ = newMetric
		}

		if response.NextToken == "" {
			break
		}

		request.NextToken = response.NextToken
	}

}

func TestSvr(t *testing.T) {

	ag := NewAgent()

	data, err := ioutil.ReadFile("./test.conf")
	if err != nil {
		log.Fatalf("%s", err)
	}

	err = toml.Unmarshal(data, ag)
	if err != nil {
		log.Fatalf("%s", err)
	}

	ag.Run()

}
