package aliyuncms

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/influxdata/toml"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"

	"github.com/influxdata/telegraf/metric"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
)

func apiClient() *cms.Client {
	client, err := cms.NewClientWithAccessKey("cn-hangzhou", "LTAI4FwpUNoPEFj7kQScDrDE", "CI8Lzj22RODi3L79jzMmR3gKjMe3YG")
	if err != nil {
		log.Fatalf("%s", err)
	}
	return client
}

func TestLoadConfig(t *testing.T) {

	cfgData := `
# ## [[cms]] 块可以有多个， 每个 [[cms]] 块代表一个账号.
[[cms]]

	# ##(required) 阿里云API访问 access key及区域， 至少拥有 "只读访问云监控（CloudMonitor）"的权限.
	access_key_id = 'aa'
	access_key_secret = 'aa'
	region_id = 'cn-hangzhou'

	# ##(optional) 全局的采集间隔，每个指标可以单独配置，默认5分钟.
	interval = '5m'

	# ##(optional) 阿里云监控项数据可能在当前采集时间点之后才可用，配置此项用于获取该延迟时间段的数据，如果设置为0可能导致数据不完整.  
	# ## 不同的指标可能有不同的延迟时间, 默认为5分钟, 你可以根据使用中的实际采集情况调整该值.
	delay = '5m'

	# ##(required) [[cms.project]] 块可以有多个，每个代表一个云产品.
	[[cms.project]]
	#	##(required) 云产品命名空间，可参考: https://help.aliyun.com/document_detail/28619.html?spm=a2c4g.11186623.6.690.9dbe5679uFUe3w
	name='acs_ecs_dashboard'

	# ##(required) 配置采集指标
	[cms.project.metrics]

	# ##(required) 指定采集当前产品下的哪些指标
	# ## 每个产品支持的指标可参考: See: https://help.aliyun.com/document_detail/28619.html?spm=a2c4g.11186623.2.11.6ac47694AjhHt4
	names = [
		'CPUUtilization',
	]

	# ##(optional) 定义每个指标的采集行为
	[[cms.project.metrics.property]]

	# ##(required) 指定设置哪个指标的属性, 必须在上面配置的指标名列表中, 否则忽略.
	# ## 可以使用 * 来全局配置当前project的指标采集行为.
	name = "CPUUtilization"
	
	# ##(optional) 指标采样周期, 单位为秒.
	# ## 指标项的Period可参考: See: https://help.aliyun.com/document_detail/28619.html?spm=a2c4g.11186623.2.11.6ac47694AjhHt4
	# ## 如果没有配置或配置了不支持的period，默认会使用该监控项支持的最低采样周期(一般为60s).
	period = 60

	# ##(optional) 可单独配置指标的采集间隔, 没有则使用全局配置
	interval = '5m'

	# ##(optional) 配置采集维度.
	dimensions = '''
	  [
		{"instanceId":"i-bp15wj5w33t8vfxi****"},
		{"instanceId":"i-bp1bq3x84ko4ct6x****"}
		]
		'''
`

	var cfg CmsAgent

	err := toml.Unmarshal([]byte(cfgData), &cfg)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		log.Printf("ok")
	}
}

//查询对应产品可以获取哪些监控项
func TestGetMetricMeta(t *testing.T) {

	request := cms.CreateDescribeMetricMetaListRequest()
	request.Scheme = "https"
	request.Namespace = `acs_kvstore`
	//request.MetricName = `LatencyByConnectionRegion`
	request.PageSize = requests.NewInteger(200)

	response, err := apiClient().DescribeMetricMetaList(request)
	if err != nil {
		t.Error(err)
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
			log.Printf("%s", internal.Metric2InfluxLine(newMetric))
		}

		if response.NextToken == "" {
			break
		}

		request.NextToken = response.NextToken
	}

}

func TestSvr(t *testing.T) {

	ag := NewAgent()

	data, err := ioutil.ReadFile("./aliyuncms.toml")
	if err != nil {
		log.Fatalf("%s", err)
	}

	err = toml.Unmarshal(data, ag)
	if err != nil {
		log.Fatalf("%s", err)
	}

	ag.Start(nil)

	time.Sleep(time.Hour)
}
