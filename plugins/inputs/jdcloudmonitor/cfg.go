package jdcloudmonitor

import (
	"context"
	"time"

	"golang.org/x/time/rate"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"

	jcwclient "github.com/jdcloud-api/jdcloud-sdk-go/services/monitor/client"
)

const (
	sampleConfig = `
# ##(required)
#[[inputs.jdcloudmonitor]]
# ##(required)
#region_id = 'cn-south-1'

# ##(required)
#access_key_id = ''

# ##(required)
#access_key_secret = ''

# ##(optional) collect interval
#interval = '5m'

# ##(optional) default is 1min, should not more then interval
#delay = '1m'

# ##(optional) custom tags
#[inputs.jdcloudmonitor.tags]
#key1 = 'val1'

# ##(required)
#[[inputs.jdcloudmonitor.services]]
# ##(required) service name
#name = 'vm'

# ##(required) resource id
#resource_id = 'i-dcnxfxxxxx'

# ##(required) metric names
#metrics = ['cpu_util', 'memory.usage']

# ##(optional)
#[[inputs.jdcloudmonitor.services.property]]
#name = 'memory.usage'
#aggregate_type = 'max'
#interval = '10m'

# ##(optional)
#[inputs.jdcloudmonitor.services.property.tags]
#key1 = 'val1'
`
)

type (
	Property struct {
		Name          string            `toml:"name"`
		AggregateType string            `toml:"aggregate_type"`
		Interval      datakit.Duration  `toml:"interval"`
		Tags          map[string]string `toml:"tags,omitempty"`
	}

	Service struct {
		Name        string      `toml:"name"`
		ResourceID  string      `toml:"resource_id"`
		MetricNames []string    `toml:"metrics"`
		Property    []*Property `toml:"property"`

		globalMetricProperty *Property
	}

	agent struct {
		RegionID        string            `toml:"region_id"`
		AccessKeyID     string            `toml:"access_key_id"`
		AccessKeySecret string            `toml:"access_key_secret"`
		Interval        datakit.Duration  `toml:"interval"`
		Delay           datakit.Duration  `toml:"delay"`
		Tags            map[string]string `toml:"tags,omitempty"`
		Services        []*Service        `toml:"services"`

		client *jcwclient.MonitorClient

		limiter *rate.Limiter

		ctx       context.Context
		cancelFun context.CancelFunc

		reqs []*metricsRequest

		debugMode bool
	}

	metricsRequest struct {
		tags map[string]string

		servicename string
		resourceid  string
		metricname  string
		from        string
		to          string

		//每个指标可单独配置interval，默认使用全局的配置
		interval  time.Duration
		aggreType string

		lastTime time.Time
	}
)

func (p *Service) genMetricReq(metric string) (*metricsRequest, error) {

	req := &metricsRequest{
		servicename: p.Name,
		resourceid:  p.ResourceID,
		metricname:  metric,
	}

	if p.globalMetricProperty == nil {
		for _, prop := range p.Property {
			if prop.Name == "*" {
				p.globalMetricProperty = prop
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
		property = p.globalMetricProperty
	}

	if property != nil {
		req.interval = property.Interval.Duration
		req.tags = property.Tags
		req.aggreType = property.AggregateType
	}

	if req.aggreType == "" {
		req.aggreType = "avg"
	}

	return req, nil
}
