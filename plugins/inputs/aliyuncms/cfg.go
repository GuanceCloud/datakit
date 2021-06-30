package aliyuncms

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
	"golang.org/x/time/rate"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	aliyuncmsConfigSample = `
# ##(required)
#[[inputs.aliyuncms]]

 # ##(required)
 # access_key_id = ''
 # access_key_secret = ''
 # region_id = 'cn-hangzhou'

 # ##(optional) Security Token Service(STS)
 ## security_token = ''

 # ##(optional) global collect interval, default is 5min.
 ## interval = '5m'

 # ##(optional) delay collect duration
 ## delay = '5m'

 # ##(optional) custom tags
 ## [inputs.aliyuncms.tags]
 ##  key1 = "val1"
 ##  key2 = "val2"

# ##(required)
#[[inputs.aliyuncms.project]]
# ##(required) product namespace
# namespace='acs_ecs_dashboard'

# ##(optional) names of metrics, comma-separated, if empty, collect all metrics of the project
## metric = 'CPUUtilization,DiskWriteBPS'

# ##(optional)
## [[inputs.aliyuncms.project.property]]

# ##(optional) comma-separated metrics which this property will apply to, if empty, apply to all
## name = 'CPUUtilization,DiskWriteBPS'

# ##(optional) you may specify period of this metric
## period = 60

# ##(optional) collect interval of thie metric
## interval = '5m'

# ##(optional) collect filter, a json string
## dimensions = '''
##  [
##	 {"instanceId":"i-bp15wj5w33t8vf******"}
##	]
##	'''

# ##(optional) custom tags for the metrics
## [inputs.aliyuncms.project.property.tags]
## key1 = "val1"
## key2 = "val2"
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

	//deprecated
	Metric struct {
		MetricNames []string    `toml:"names"`
		Property    []*Property `toml:"property,omitempty"`
	}

	Project struct {
		Name      string `toml:"name"` //deprecated
		Namespace string `toml:"namespace"`

		MetricNames string `toml:"metric,omitempty"`

		MetricName string `toml:"metric_name"` //deprecated

		Metrics *Metric `toml:"metrics"` //deprecated

		Property []*Property `toml:"property,omitempty"`

		properties map[string]*Property
	}

	CloudApiCallInfo struct {
		total   uint64
		details map[string][]uint64
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

		apiClient *cms.Client

		reqs []*MetricsRequest

		limiter *rate.Limiter

		mode string

		apiCallInfo *CloudApiCallInfo
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

		//每个指标可单独配置interval，默认使用全局的配置
		interval time.Duration

		measurementName string

		tryGetMeta int

		//当period的配置不支持时调整period
		tunePeriod bool

		tuneDimension bool

		lastTime time.Time
	}
)

func (p *Project) makeReqWrap(metricName string) *MetricsRequest {

	req := cms.CreateDescribeMetricListRequest()
	req.Scheme = "https"
	//req.RegionId = region
	req.Period = "60"
	req.MetricName = metricName
	req.Namespace = p.namespace()

	reqWrap := &MetricsRequest{
		q:               req,
		tryGetMeta:      5,
		measurementName: p.MetricName,
	}

	return reqWrap
}

func (p *Project) namespace() string {
	if p.Namespace == "" {
		return p.Name
	}
	return p.Namespace
}

func (p *Project) checkProperties() error {
	p.properties = make(map[string]*Property)

	props := p.Property
	if props == nil && p.Metrics != nil {
		props = p.Metrics.Property
	}
	for _, prop := range props {
		checkDimensions := []map[string]string{}
		if err := json.Unmarshal([]byte(prop.Dimensions), &checkDimensions); err != nil {
			err = fmt.Errorf("invalid dimension of property '%s', error: %s", prop.Name, err)
			moduleLogger.Errorf("%s", err)
			return err
		}
		prop.Dimensions = strings.Trim(prop.Dimensions, " \t\r\n")

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
					p.properties[sn] = np
				}
			} else if len(parts) == 1 {
				p.properties[prop.Name] = prop
			}
		}

	}

	return nil
}

func (p *Project) applyProperty(req *MetricsRequest) {

	metricName := req.q.MetricName

	var property *Property
	property = p.properties[metricName]
	if property == nil {
		property = p.properties["*"]
	}

	if property != nil {
		if property.Period > 0 {
			req.q.Period = strconv.FormatInt(int64(property.Period), 10)
		}
		if property.Interval.Duration != 0 {
			req.interval = property.Interval.Duration
		}
		req.tags = property.Tags
		req.q.Dimensions = property.Dimensions
	}

}

func (a *CMS) isDebug() bool {
	return a.mode == "debug"
}

func (i *CloudApiCallInfo) Inc(apiname string, fail bool) {
	idx := 0
	if fail {
		idx++
	}
	p := i.details[apiname]
	if p == nil {
		i.details[apiname] = []uint64{0, 0}
		p = i.details[apiname]
	}
	p[idx] = p[idx] + 1
}

func (i *CloudApiCallInfo) String() string {
	s := fmt.Sprintf("Total: %v\n", i.total)
	for k, v := range i.details {
		s += fmt.Sprintf("%s=%v, %v\n", k, v[0], v[1])
	}
	return s
}
