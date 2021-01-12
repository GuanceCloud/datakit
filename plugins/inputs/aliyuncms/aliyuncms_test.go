package aliyuncms

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	"github.com/influxdata/telegraf/metric"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/resourcemanager"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/tag"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
)

func TestInput(t *testing.T) {

	ag := loadCfg("test.conf").(*CMS)
	ag.mode = "debug"
	ag.Run()
}

func TestGetMetricMeta(t *testing.T) {

	ag := loadCfg("test.conf").(*CMS)

	client, err := cms.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
	if err != nil {
		log.Fatalf("%s", err)
	}

	request := cms.CreateDescribeMetricMetaListRequest()
	request.Scheme = "https"
	request.Namespace = `acs_ens` // `acs_ecs_dashboard`
	pageSize := 1000
	request.PageSize = requests.NewInteger(pageSize)

	for {
		response, err := client.DescribeMetricMetaList(request)
		if err != nil {
			eTyp := reflect.TypeOf(err)
			log.Printf("error type: %v, kind: %v", eTyp.String(), eTyp.Kind())

			if se, ok := err.(*errors.ServerError); ok {
				log.Printf("error code: %v, %v", se.ErrorCode(), se.Message())

			}
			t.Errorf("error: %s", err)
			return
		}

		log.Printf("respcode: %v, %v", response.Code, response)

		thisCount := len(response.Resources.Resource)
		log.Printf("get=%d", thisCount)

		if response.Resources.Resource != nil {
			for _, res := range response.Resources.Resource {
				//有些指标的Statistics返回为空
				log.Printf("%s: Periods=%s Statistics=%s, Dimension=%s, Unit=%v", res.MetricName, res.Periods, res.Statistics, res.Dimensions, res.Unit)
			}
		}

		if thisCount < pageSize {
			break
		}
	}
}

func TestDescribeProjectMeta(t *testing.T) {

	ag := loadCfg("test.conf").(*CMS)

	client, err := cms.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
	if err != nil {
		log.Fatalf("%s", err)
	}

	request := cms.CreateDescribeProjectMetaRequest()
	request.Scheme = "https"
	pageSize := 1000
	request.PageSize = requests.NewInteger(pageSize)

	total := ""

	for {
		response, err := client.DescribeProjectMeta(request)
		if err != nil {
			eTyp := reflect.TypeOf(err)
			log.Printf("error type: %v, kind: %v", eTyp.String(), eTyp.Kind())

			if se, ok := err.(*errors.ServerError); ok {
				log.Printf("error code: %v, %v", se.ErrorCode(), se.Message())

			}
			t.Errorf("error: %s", err)
			return
		}

		//log.Printf("respcode: %v, %v", response.Code, response)

		if total == "" {
			log.Printf("total=%v", response.Total)
			total = response.Total
		}

		thisCount := len(response.Resources.Resource)
		log.Printf("get=%d", thisCount)

		for _, res := range response.Resources.Resource {
			var labels []map[string]string
			if err := json.Unmarshal([]byte(res.Labels), &labels); err == nil {
				for _, l := range labels {
					if nv, ok := l["name"]; ok && nv == "product" {
						log.Printf("%s - (%s) - %s", res.Namespace, res.Description, l["value"])
						break
					}
				}
			}
		}

		if thisCount < pageSize {
			break
		}
	}
}

func TestListResources(t *testing.T) {

	ag := loadCfg("test.conf").(*CMS)

	client, err := resourcemanager.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)

	request := resourcemanager.CreateListResourcesRequest()
	request.Scheme = "https"
	request.PageSize = requests.NewInteger(100)

	response, err := client.ListResources(request)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	for _, res := range response.Resources.Resource {
		log.Printf("%s-%s", res.Service, res.ResourceType)
	}
}

func TestListTagResources(t *testing.T) {

	ag := loadCfg("test.conf").(*CMS)

	client, err := tag.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)

	request := tag.CreateListTagResourcesRequest()
	request.Scheme = "https"
	request.PageSize = requests.NewInteger(100)

	request.Category = "Custom"

	response, err := client.ListTagResources(request)
	if err != nil {
		log.Fatalln(err.Error())
	}
	fmt.Printf("response is %#v\n", response)
}

func TestGetMetricList(t *testing.T) {

	ag := loadCfg("test.conf").(*CMS)

	request := cms.CreateDescribeMetricListRequest()
	request.Scheme = "https"
	request.RegionId = ag.RegionID
	request.MetricName = "net_rx.rate"
	request.Namespace = "acs_nat_gateway"
	request.Period = "60"
	//request.StartTime = "1588156800000"
	//request.EndTime = "1588157100000"
	//request.StartTime = "1585209420000"
	//request.EndTime = "1585209600000"

	//request.Dimensions = `[{"instanceId":"i-bp1bnmpvz2tuu7odx6cx"}]`
	request.NextToken = ""

	apiClient, _ := cms.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)

	for {
		response, err := apiClient.DescribeMetricList(request)
		if err != nil {
			t.Errorf("DescribeMetricList error, %s", err)
			return
		}

		log.Printf("response: %v", response)

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

func loadCfg(file string) inputs.Input {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("read config file failed: %s", err.Error())
		return nil
	}

	tbl, err := toml.Parse(data)
	if err != nil {
		log.Fatalf("parse toml file failed, %s", err)
		return nil
	}

	creator := func() inputs.Input {
		return NewAgent()
	}

	for field, node := range tbl.Fields {

		switch field {
		case "inputs":
			stbl, ok := node.(*ast.Table)
			if !ok {
				log.Fatalf("ignore bad toml node")
			} else {
				for inputName, v := range stbl.Fields {
					input, err := tryUnmarshal(v, inputName, creator)
					if err != nil {
						log.Fatalf("unmarshal input %s failed: %s", inputName, err.Error())
						return nil
					}
					return input
				}
			}

		default: // compatible with old version: no [[inputs.xxx]] header
			input, err := tryUnmarshal(node, "aa", creator)
			if err != nil {
				log.Fatalf("unmarshal input failed: %s", err.Error())
				return nil
			}
			return input
		}
	}

	return nil
}

func tryUnmarshal(tbl interface{}, name string, creator inputs.Creator) (inputs.Input, error) {

	tbls := []*ast.Table{}

	switch t := tbl.(type) {
	case []*ast.Table:
		tbls = tbl.([]*ast.Table)
	case *ast.Table:
		tbls = append(tbls, tbl.(*ast.Table))
	default:
		err := fmt.Errorf("invalid toml format: %v", t)
		return nil, err
	}

	for _, t := range tbls {
		input := creator()

		err := toml.UnmarshalTable(t, input)
		if err != nil {
			log.Fatalf("toml unmarshal failed: %v", err)
			return nil, err
		}
		return input, nil

	}
	return nil, nil
}
