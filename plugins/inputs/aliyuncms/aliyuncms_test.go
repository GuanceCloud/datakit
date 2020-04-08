package aliyuncms

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"testing"
	"time"

	"golang.org/x/time/rate"

	"github.com/influxdata/toml"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	//"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/limiter"

	//"github.com/siddontang/go-log/log"

	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/selfstat"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials/providers"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
)

func limitFn(n int) {
	log.Printf("limitFn-%d", n)
}

func dumpStat(stat selfstat.Stat) {

	name := stat.Name()
	tags := stat.Tags()
	fields := map[string]interface{}{
		stat.FieldName(): stat.Get(),
	}
	metric, err := metric.New(name, tags, fields, time.Now())
	if err != nil {
		log.Fatalf("%s", err)
	}
	line := internal.Metric2InfluxLine(metric)
	log.Printf("%s", line)
}

func TestLimit(t *testing.T) {

	limit := rate.Every(40 * time.Millisecond)
	limiter := rate.NewLimiter(limit, 1)
	_ = limiter

	ctx, cancelFun := context.WithCancel(context.Background())
	_ = cancelFun
	_ = ctx

	for {

		t := time.Now()

		for i := 0; i < 60; i++ {
			limiter.Wait(ctx)
			limitFn(i)
		}

		useage := time.Now().Sub(t)
		if useage < batchInterval {
			remain := batchInterval - useage
			log.Printf("remain: %v", remain)
			time.Sleep(remain)
		}
	}
}

func TestConfig(t *testing.T) {

	var cfg CmsAgent

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

	//client, err := cms.NewClientWithAccessKey("cn-hangzhou", "LTAIlsWpTrg1vUf4", "dy5lQzWpU17RDNHGCj84LBDhoU9LVU")

	client, err := cms.NewClientWithAccessKey("cn-hangzhou", "LTAI4FwpUNoPEFj7kQScDrDE", "CI8Lzj22RODi3L79jzMmR3gKjMe3YG")

	//namespace := "acs_ecs_dashboard"
	//metricname := "IOPSUsage"

	namespace := "acs_kubernetes"
	metricname := "group.disk.io_read_bytes"
	_ = metricname

	request := cms.CreateDescribeMetricMetaListRequest()
	request.Scheme = "https"

	request.Namespace = namespace
	request.MetricName = metricname
	//request.PageSize = requests.NewInteger(100)

	response, err := client.DescribeMetricMetaList(request)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}

	if response.Resources.Resource != nil {
		for _, res := range response.Resources.Resource {
			fmt.Printf("%s: Periods=%s Statistics=%s Dimensions=%s\n", res.MetricName, res.Periods, res.Statistics, res.Dimensions)
		}
	}
}

func TestGetMetricMeta(t *testing.T) {

	client, err := cms.NewClientWithAccessKey("cn-hangzhou", "LTAI4FwpUNoPEFj7kQScDrDE", "CI8Lzj22RODi3L79jzMmR3gKjMe3YG")

	namespace := "acs_kubernetes"

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

func TestGetMetricList(t *testing.T) {

	//get 96 datapoints: Namespace=acs_kubernetes, MetricName=group.cpu.limit, Period=60, StartTime=1585210020000(2020-03-26 16:07:00 +0800 CST), EndTime=1585210320000(2020-03-26 16:12:00 +0800 CST), Dimensions=

	//get 172 datapoints: Namespace=acs_ecs_dashboard, MetricName=cpu_total, Period=60, StartTime=1585211520000(2020-03-26 16:32:00 +0800 CST), EndTime=1585211820000(2020-03-26 16:37:00 +0800 CST), Dimensions=

	//2020-03-26T08:37:21Z D! [aliyuncms] get 172 datapoints: Namespace=acs_ecs_dashboard, MetricName=memory_usedspace, Period=60, StartTime=1585211520000(2020-03-26 16:32:00 +0800 CST), EndTime=1585211820000(2020-03-26 16:37:00 +0800 CST), Dimensions=

	//client, _ := cms.NewClientWithAccessKey("cn-hangzhou", "LTAI4FwpUNoPEFj7kQScDrDE", "CI8Lzj22RODi3L79jzMmR3gKjMe3YG")

	configuration := &providers.Configuration{
		AccessKeyID:     "LTAIaB2ZMYy4Dej9",
		AccessKeySecret: `pixGuiJail10JSBZTzuaOJIw8N2pw7`,
	}
	credentialProviders := []providers.Provider{
		providers.NewConfigurationCredentialProvider(configuration),
		providers.NewEnvCredentialProvider(),
		providers.NewInstanceMetadataProvider(),
	}
	credential, err := providers.NewChainProvider(credentialProviders).Retrieve()
	if err != nil {
		t.Errorf("failed to retrieve credential")
	}

	client, err := cms.NewClientWithOptions(`cn-hangzhou`, sdk.NewConfig(), credential)
	if err != nil {
		t.Errorf("failed to create cms client: %v", err)
	}

	request := cms.CreateDescribeMetricListRequest()
	request.Scheme = "https"
	request.RegionId = "cn-hangzhou"
	request.MetricName = "TrafficRXNew"
	request.Namespace = "acs_slb_dashboard"
	request.Period = "60"
	request.StartTime = "1585883160000"
	request.EndTime = "1585883760000"
	//request.StartTime = "1585209420000"
	//request.EndTime = "1585209600000"

	request.Dimensions = "" // `[{"instanceId":"i-bp1dsyh39swucxotofde"},{}]`

	for i := 0; i < 1; i++ {

		request.NextToken = ""

		response, err := client.DescribeMetricList(request)
		if err != nil {
			t.Errorf("xxx %s", err)
			return
		}

		var datapoints []map[string]interface{}
		if err = json.Unmarshal([]byte(response.Datapoints), &datapoints); err != nil {
			t.Errorf("failed to decode response datapoints: %v", err)
		}

		log.Printf("Count: %v", len(datapoints))

		continue

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

			for _, k := range dms {
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
			//log.Printf("%s", internal.Metric2InfluxLine(newMetric))

			//AddFields(formatMeasurement(req.q.Namespace), fields, tags, tm)
		}

	}

	// errCtx := map[string]string{
	// 	"Scheme":     request.Scheme,
	// 	"MetricName": request.MetricName,
	// 	"Namespace":  request.Namespace,
	// 	"Dimensions": request.Dimensions,
	// 	"Period":     request.Period,
	// 	"StartTime":  request.StartTime,
	// 	"EndTime":    request.EndTime,
	// 	"NextToken":  request.NextToken,
	// 	"RegionId":   request.RegionId,
	// 	"Domain":     request.Domain,
	// 	"RequestId":  response.RequestId,
	// 	"Code":       response.Code,
	// 	"Message":    response.Message,
	// }

	// jstr, _ := json.Marshal(errCtx)

	// log.Printf("errCtx: %s", string(jstr))

	//fmt.Printf("**%s\n", response.String())

}

func TestSvr(t *testing.T) {

	var cfg CmsAgent

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
