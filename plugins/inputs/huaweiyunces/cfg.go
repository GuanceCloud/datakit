package huaweiyunces

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/time/rate"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/sdk/huaweicloud"
)

const (
	sampleConfig = `
#[[inputs.huaweiyunces]]
# ##(required) the following 4 configurations are required for authentication
#access_key_id = ''
#access_key_secret = ''
#endpoint = ''
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

# ##(required) you can specify up to 3 dimensions
#[inputs.huaweiyunces.namespace.property.dimensions]
#instance_id = 'b5d7b7a3-681d-4c08-8e32-f14******'
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
		ProjectID       string            `toml:"projectid"`
		Interval        datakit.Duration  `toml:"interval"`
		Delay           datakit.Duration  `toml:"delay"`
		Tags            map[string]string `toml:"tags,omitempty"`
		Namespace       []*Namespace      `toml:"namespace"`

		client *huaweicloud.HWClient

		limiter *rate.Limiter

		ctx       context.Context
		cancelFun context.CancelFunc

		reqs []*metricsRequest
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

func (p *Namespace) genMetricReq(metric string) (*metricsRequest, error) {

	req := &metricsRequest{
		namespace:     p.Name,
		metricname:    metric,
		tryGetMeta:    5,
		metricSetName: p.MetricSetName,
	}

	if p.globalMetricProperty == nil {
		for _, prop := range p.Property {
			if prop.Name == "*" {
				p.globalMetricProperty = prop
				if prop.Dimensions == nil {
					return nil, fmt.Errorf("invalid property, dimensions cannot be empty")
				}
				break
			}
		}
	}

	var property *Property

	for _, prop := range p.Property {
		if prop.Name == "*" {
			continue
		}
		if prop.Name == metric {
			property = prop
			break
		}
	}

	if property == nil {
		if p.globalMetricProperty == nil {
			return nil, fmt.Errorf("no property found for %s", metric)
		}
		property = p.globalMetricProperty
	} else {
		if property.Dimensions == nil {
			if p.globalMetricProperty != nil && p.globalMetricProperty.Dimensions != nil {
				property.Dimensions = p.globalMetricProperty.Dimensions
			} else {
				return nil, fmt.Errorf("invalid property, dimensions cannot be empty")
			}
		}
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

	return req, nil
}
