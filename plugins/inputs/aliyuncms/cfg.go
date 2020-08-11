package aliyuncms

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	aliyuncmsConfigSample = `
#[[inputs.aliyuncms]]

 # ##(required)
 #access_key_id = ''
 #access_key_secret = ''
 #region_id = 'cn-hangzhou'

 # ##(optional)（Security Token Service，STS）
 #security_token = ''

 # ##(optional) global collect interval, default is 5min.
 #interval = '5m'

 # ##(optional) delay collect duration
 #delay = '5m'

 #[inputs.aliyuncms.tags]
 #key1 = "val1"
 #key2 = "val2"

 # ##(required)
 #[[inputs.aliyuncms.project]]
  #	##(required) product namespace
  #name='acs_ecs_dashboard'

  # ##(optional) custom metric name，default is "aliyuncms_<name>"
  #metric_name=''

  # ##(required)
  #[inputs.aliyuncms.project.metrics]

   # ##(required)
   # ## names of metrics
   #names = [
   #	'CPUUtilization',
   #]

   # ##(optional)
   #[[inputs.aliyuncms.project.metrics.property]]

	# ##(required) you can use * to apply to all metrics of this project
	#name = "CPUUtilization"

	# ##(optional) you may specify period of this metric
	#period = 60

	# ##(optional) collect interval of thie metric
	#interval = '5m'

	# ##(optional) collect filter, a json string
	#dimensions = '''
    #  [
    #	{"instanceId":"i-bp15wj5w33t8vf******"}
    #	]
	#	'''

	# ##(optional) custom tags
	#[inputs.aliyuncms.project.metrics.property.tags]
	#key1 = "val1"
	#key2 = "val2"
`
)

type (
	Dimension struct {
		Name   string `toml:"name"`
		Value  string `toml:"value"`
		Period int    `toml:"period"`
	}

	Property struct {
		Name       string           `toml:"name"`
		Period     int              `toml:"period"`
		Interval   datakit.Duration `toml:"interval"`
		Dimensions string           `toml:"dimensions"`

		Tags map[string]string `toml:"tags,omitempty"`
	}

	Metric struct {
		MetricNames []string    `toml:"names"`
		Property    []*Property `toml:"property,omitempty"`

		//兼容老的配置
		Dimensions []*Dimension `toml:"dimensions,omitempty"`
	}

	Project struct {
		Name       string `toml:"name"`
		MetricName string `toml:"metric_name"`

		//兼容老的配置
		InstanceIDs []string `toml:"instanceIds,omitempty"`

		Metrics *Metric `toml:"metrics"`

		//该Project下的全局指标采集属性
		globalMetricProperty *Property
	}

	CMS struct {
		RegionID        string            `toml:"region_id"`
		AccessKeyID     string            `toml:"access_key_id"`
		AccessKeySecret string            `toml:"access_key_secret"`
		SecurityToken   string            `toml:"security_token"`
		Interval        datakit.Duration  `toml:"interval"`
		Delay           datakit.Duration  `toml:"delay"`
		Project         []*Project        `toml:"project"`
		Tags            map[string]string `toml:"tags,omitempty"`

		ctx       context.Context
		cancelFun context.CancelFunc
	}

	MetricMeta struct {
		//指标支持的采样周期
		Periods []int64

		//指标支持的统计方式
		Statistics []string

		//指标支持的维度
		Dimensions []string

		//指标的描述
		Description string

		//指标单位
		Unit string

		metricName string
	}

	MetricsRequest struct {
		q *cms.DescribeMetricListRequest

		tags map[string]string

		meta *MetricMeta

		metricSetName string

		//每个指标可单独配置interval，默认使用全局的配置
		interval time.Duration

		tryGetMeta int

		//period的配置不支持时调整period
		tunePeriod bool

		tuneDimension bool

		lastTime time.Time
	}
)

func (p *Project) genMetricReq(metric string, region string) (*MetricsRequest, error) {

	req := cms.CreateDescribeMetricListRequest()
	req.Scheme = "https"
	req.RegionId = region
	req.Period = "60"
	req.MetricName = metric
	req.Namespace = p.Name

	var interval time.Duration

	var metricTags map[string]string

	if p.Metrics.Dimensions != nil || len(p.InstanceIDs) > 0 { //兼容老的配置

		dimensions := []map[string]string{}

		var dimension *Dimension

		for _, d := range p.Metrics.Dimensions {
			if d.Name == metric {
				dimension = d
				if d.Period > 0 {
					req.Period = strconv.FormatInt(int64(d.Period), 10)
				}
				break
			}
		}

		if dimension != nil && dimension.Value != "" {
			if err := json.Unmarshal([]byte(dimension.Value), &dimensions); err != nil {
				return nil, fmt.Errorf("invalid dimension(%s.%s): %s, %s", p.Name, metric, dimension.Value, err)
			}
		}

		bHaveSetID := false
		for _, m := range dimensions {
			if _, ok := m["instanceId"]; ok {
				bHaveSetID = true
				break
			}
		}

		//如果dimension中没有配置instanceId，而全局配置了，则添加进去
		if !bHaveSetID && len(p.InstanceIDs) > 0 {
			for _, id := range p.InstanceIDs {
				dimensions = append(dimensions, map[string]string{
					"instanceId": id})
			}
		}

		if len(dimensions) > 0 {
			js, err := json.Marshal(&dimensions)
			if err != nil {
				return nil, err
			}
			req.Dimensions = string(js)
		}

	} else {

		if p.globalMetricProperty == nil {
			for _, prop := range p.Metrics.Property {
				if prop.Name == "*" {
					p.globalMetricProperty = prop
					if prop.Dimensions != "" {
						checkDimensions := []map[string]string{}
						if err := json.Unmarshal([]byte(prop.Dimensions), &checkDimensions); err != nil {
							return nil, fmt.Errorf("invalid dimension(%s): %s, %s", metric, prop.Dimensions, err)
						}
						p.globalMetricProperty.Dimensions = strings.Trim(prop.Dimensions, " \t\r\n")
					}
					break
				}
			}
		}

		var property *Property

		for _, prop := range p.Metrics.Property {
			if prop.Name == "*" {
				continue
			}
			if prop.Name == metric {
				property = prop
				if prop.Period > 0 {
					req.Period = strconv.FormatInt(int64(prop.Period), 10)
				}
				if prop.Interval.Duration != 0 {
					interval = prop.Interval.Duration
				}
				metricTags = property.Tags
				break
			}
		}

		if property == nil && p.globalMetricProperty != nil {
			if p.globalMetricProperty.Period > 0 {
				req.Period = strconv.FormatInt(int64(p.globalMetricProperty.Period), 10)
			}
			interval = p.globalMetricProperty.Interval.Duration
			metricTags = p.globalMetricProperty.Tags
		}

		if property != nil && property.Dimensions != "" {
			//检查配置是否正确
			checkDimensions := []map[string]string{}
			if err := json.Unmarshal([]byte(property.Dimensions), &checkDimensions); err != nil {
				return nil, fmt.Errorf("invalid dimension(%s): %s, %s", metric, property.Dimensions, err)
			}
			req.Dimensions = strings.Trim(property.Dimensions, " \t\r\n")
		}

		if req.Dimensions == "" && p.globalMetricProperty != nil {
			req.Dimensions = p.globalMetricProperty.Dimensions
		}
	}

	reqWrap := &MetricsRequest{
		q:             req,
		tags:          metricTags,
		interval:      interval,
		tryGetMeta:    5,
		metricSetName: p.MetricName,
	}

	return reqWrap, nil
}

func errCtxMetricList(req *MetricsRequest) map[string]string {

	return map[string]string{
		"Scheme":     req.q.Scheme,
		"MetricName": req.q.MetricName,
		"Namespace":  req.q.Namespace,
		"Dimensions": req.q.Dimensions,
		"Period":     req.q.Period,
		"StartTime":  req.q.StartTime,
		"EndTime":    req.q.EndTime,
		"NextToken":  req.q.NextToken,
		"RegionId":   req.q.RegionId,
		"Domain":     req.q.Domain,
	}
}

func errCtxMetricMeta(req *cms.DescribeMetricMetaListRequest) map[string]string {
	return map[string]string{
		"Scheme":     req.Scheme,
		"MetricName": req.MetricName,
		"Namespace":  req.Namespace,
		"RegionId":   req.RegionId,
		"Domain":     req.Domain,
	}
}
