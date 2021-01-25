package huaweiyunces

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/time/rate"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	ces "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1"
)

const (
	sampleConfig = `
#[[inputs.huaweiyunces]]
# ##(required) the following 4 configurations are required for authentication
#access_key_id = ''
#access_key_secret = ''
#region = ''
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

# ##(optional) default is average, support values are: max, min, average, sum, variance
#filter = ''

# ##(optional) if not set, use the global interval
#interval = '5m'

# ##(required) you can specify up to 4 dimensions
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

		properties map[string]*Property
	}

	agent struct {
		AccessKeyID     string            `toml:"access_key_id"`
		AccessKeySecret string            `toml:"access_key_secret"`
		EndPoint        string            `toml:"endpoint"` //deprated
		RegionID        string            `toml:"region"`
		ProjectID       string            `toml:"projectid"`
		Interval        datakit.Duration  `toml:"interval"`
		Delay           datakit.Duration  `toml:"delay"`
		Tags            map[string]string `toml:"tags,omitempty"`
		Namespace       []*Namespace      `toml:"namespace"`

		ecsInstanceIDs []string

		client *ces.CesClient

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

		lastTime time.Time
	}
)

func (r *metricsRequest) copy() *metricsRequest {
	var m metricsRequest
	m = *r
	return &m
}

func (ag *agent) isTestOnce() bool {
	return ag.mode == "test"
}

func (ag *agent) isDebug() bool {
	return ag.mode == "debug"
}

func (p *Namespace) checkProperties() error {
	p.properties = make(map[string]*Property)

	for _, prop := range p.Property {

		if prop.Dimensions == nil && p.Name != "SYS.ECS" {
			err := fmt.Errorf("invalid property, dimensions cannot be empty")
			moduleLogger.Errorf("%s", err)
			return err
		}

		if prop.Name == "" {
			p.properties["*"] = prop
		} else {
			parts := strings.Split(prop.Name, ",")
			if len(parts) > 1 {
				for _, sn := range parts {
					sn = strings.TrimSpace(sn)
					np := &Property{}
					np.Dimensions = prop.Dimensions
					np.Interval = prop.Interval
					np.Period = prop.Period
					np.Name = sn
					np.Tags = prop.Tags
					np.Filter = prop.Filter
					p.properties[sn] = np
				}
			} else if len(parts) == 1 {
				p.properties[prop.Name] = prop
			}
		}

	}

	return nil
}

func (p *Namespace) applyProperty(req *metricsRequest, instanceIDs []string) (adds []*metricsRequest) {

	ecs := (p.Name == "SYS.ECS")

	metricName := req.metricname

	var property *Property
	property = p.properties[metricName]
	if property == nil {
		property = p.properties["*"]
	}

	if property == nil {
		if ecs {
			property = &Property{
				Name: metricName,
			}
		} else {
			return nil
		}
	}

	if property.Period > 0 {
		req.period = property.Period
	}
	if property.Filter != "" {
		req.filter = property.Filter
	}
	if property.Interval.Duration != 0 {
		req.interval = property.Interval.Duration
	}
	req.tags = property.Tags

	if len(property.Dimensions) == 0 && ecs {
		//如果是ecs且没有设置dimension，则默认拿所有instance
		for idx, inst := range instanceIDs {
			if idx == 0 {
				req.dimensoions = []*Dimension{
					&Dimension{
						Name:  "instance_id",
						Value: inst,
					},
				}
			} else {
				newreq := req.copy()
				newreq.dimensoions = []*Dimension{
					&Dimension{
						Name:  "instance_id",
						Value: inst,
					},
				}
				adds = append(adds, newreq)
			}
		}

	} else {
		dims := []*Dimension{}
		for k, v := range property.Dimensions {
			dim := &Dimension{
				Name:  k,
				Value: v,
			}
			dims = append(dims, dim)
		}
		req.dimensoions = dims
	}

	return
}

func (p *Namespace) genMetricReq(metric string) *metricsRequest {

	req := &metricsRequest{
		namespace:     p.Name,
		metricname:    metric,
		metricSetName: p.MetricSetName,
		filter:        "average",
		period:        300,
	}

	return req
}
