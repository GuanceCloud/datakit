package huaweiyunces

import (
	"context"
	"time"

	"golang.org/x/time/rate"

	"fmt"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/sdk/huaweicloud"
)

const (
	sampleConfig = `
#[[inputs.huaweiyunces]]
# ##(required) the following 4 configurations are required for authentication
#access_key_id = ''
#access_key_secret = ''
#endpoint = ''
#ecs_endpoint = ''
#projectid = ''

# ##(optional) default is 5min
#interval = '5m'

# ##(optional) default is 1min, should not more then interval
#delay = '1m'

# ##(optional) custom tags
#[inputs.huaweiyunces.tags]
#key1 = 'val1'

# ##(required)
#[[inputs.huaweiyunces.namespace]]

# ##(required) namespace's name
#name = 'SYS.ECS'

# ##(optional) metric set name, default is 'huaweiyunces_{namespace}'
#metric_set_name = ''

# ##(required) metric names
#metric_names = ['cpu_util', 'disk_write_bytes_rate']

# ##(required) you must specify for each metric, or you can set a gloabl configuration from all metrics
#[[inputs.huaweiyunces.namespace.property]]

# ##(required)
#name = '*'

# ##(optional) defalt is 300
#period = 0

# ##(optional) default is average
#filter = ''

# ##(optional) if not set, use the global interval
#interval = '5m'

`
)

type (
	Dimension struct {
		Name  string `toml:"name" json:"name"`
		Value string `toml:"value" json:"value"`
	}

	Property struct {
		Name       string            `toml:"name"`
		Period     int               `toml:"period"`
		Filter     string            `toml:"filter"`
		Interval   datakit.Duration  `toml:"interval"`
		Dimensions map[string]string `toml:"dimensions"`
		Tags       map[string]string `toml:"tags,omitempty"`
	}

	Namespace struct {
		Name          string      `toml:"name"`
		MetricSetName string      `toml:"metric_set_name"`
		MetricNames   []string    `toml:"metric_names"`
		Property      []*Property `toml:"property"`

		globalMetricProperty *Property
	}

	agent struct {
		AccessKeyID     string            `toml:"access_key_id"`
		AccessKeySecret string            `toml:"access_key_secret"`
		EndPoint        string            `toml:"endpoint"`
		EcsEndPoint     string            `toml:"ecs_endpoint"`
		ProjectID       string            `toml:"projectid"`
		Interval        datakit.Duration  `toml:"interval"`
		Delay           datakit.Duration  `toml:"delay"`
		Tags            map[string]string `toml:"tags,omitempty"`
		Namespace       []*Namespace      `toml:"namespace"`

		client    *huaweicloud.HWClient
		ecsClient *huaweicloud.HWClient

		limiter *rate.Limiter

		ctx       context.Context
		cancelFun context.CancelFunc

		reqs []*metricsRequest

		mode string

		testResult *inputs.TestResult
		testError  error
	}

	metricsRequest struct {
		tags map[string]string

		//meta *MetricMeta

		namespace   string
		metricname  string
		period      int
		filter      string
		from        int64
		to          int64
		dimensoions []*Dimension

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

func (ag *agent) isTest() bool {
	return ag.mode == "test"
}

func (ag *agent) isDebug() bool {
	return ag.mode == "debug"
}

func (p *Namespace) genMetricReq(metric string, ecsList []string, ag *agent) ([]*metricsRequest, error) {
	var request []*metricsRequest
	for _, id := range ecsList {
		req := &metricsRequest{
			namespace:     p.Name,
			metricname:    metric,
			tryGetMeta:    5,
			metricSetName: p.MetricSetName,
		}
		var property *Property
		for _, prop := range p.Property {
			prop.Dimensions = map[string]string{
				"instance_id": id,
			}
			if prop.Name == "*" {
				p.globalMetricProperty = prop
				break
			}

			if prop.Name == metric {
				property = prop
			}
		}
		if p.globalMetricProperty == nil && property == nil {
			return nil, fmt.Errorf("conf err : property empty")
		}
		if property == nil {
			property = p.globalMetricProperty
		}
		req.period = property.Period
		if req.period == 0 {
			req.period = 300
		}
		req.filter = property.Filter
		if req.filter == "" {
			req.filter = "average"
		}
		req.interval = property.Interval.Duration
		if req.interval == 0 {
			req.interval = ag.Interval.Duration
		}
		req.tags = property.Tags

		dims := []*Dimension{}
		for k, v := range property.Dimensions {
			dim := &Dimension{
				Name:  k,
				Value: v,
			}
			dims = append(dims, dim)
		}
		req.dimensoions = dims
		request = append(request, req)

	}
	return request, nil
}
