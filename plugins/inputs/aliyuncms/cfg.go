package aliyuncms

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
)

const (
	aliyuncmsConfigSample = `
# ## [[cms]] 块可以有多个， 每个 [[cms]] 块代表一个账号.
#[[cms]]

 # ##(required) 阿里云API访问 access key及区域， 至少拥有 "只读访问云监控（CloudMonitor）"的权限.
 #access_key_id = ''
 #access_key_secret = ''
 #region_id = 'cn-hangzhou'

 # ##(optional) 全局的采集间隔，每个指标可以单独配置，默认5分钟.
 #interval = '5m'

 # ##(optional) 阿里云监控项数据可能在当前采集时间点之后才可用，配置此项用于获取该延迟时间段的数据，如果设置为0可能导致数据不完整.  
 # ## 不同的指标可能有不同的延迟时间, 默认为5分钟, 你可以根据使用中的实际采集情况调整该值.
 #delay = '5m'

 # ##(required) [[cms.project]] 块可以有多个，每个代表一个云产品.
 #[[cms.project]]
  #	##(required) 云产品命名空间，可参考: https://help.aliyun.com/document_detail/28619.html?spm=a2c4g.11186623.6.690.9dbe5679uFUe3w
  #name='acs_ecs_dashboard'

  # ##(optional) 可设置指标集名称，默认使用"aliyuncms_<name>"
  #metric_name=''

  # ##(required) 配置采集指标
  #[cms.project.metrics]

   # ##(required) 指定采集当前产品下的哪些指标
   # ## 每个产品支持的指标可参考: See: https://help.aliyun.com/document_detail/28619.html?spm=a2c4g.11186623.2.11.6ac47694AjhHt4
   #names = [
   #	'CPUUtilization',
   #]

   # ##(optional) 定义每个指标的采集行为
   #[[cms.project.metrics.property]]

	# ##(required) 指定设置哪个指标的属性, 必须在上面配置的指标名列表中, 否则忽略.
	# ## 可以使用 * 来配置当前project下所有指标的采集行为.
	#name = "CPUUtilization"
	
	# ##(optional) 指标采样周期, 单位为秒.
	# ## 指标项的Period可参考: See: https://help.aliyun.com/document_detail/28619.html?spm=a2c4g.11186623.2.11.6ac47694AjhHt4
	# ## 如果没有配置或配置了不支持的period，默认会使用该监控项支持的最低采样周期(一般为60s).
	#period = 60

	# ##(optional) 可单独配置指标的采集间隔, 没有则使用全局配置
	#interval = '5m'

	# ##(optional) 配置采集维度, 是一个key-value列表的json字符串.
	# ##维度map，用于查询指定资源的监控数据. 格式为key-value键值对形式的集合，常用的key-value集合为instanceId：XXXXXX.key和value的长度为1~64个字节，超过64个字节时截取前64字节.
	# ##如果某个维度不在指标的支持范围内, 则被忽略.
	#dimensions = '''
    #  [
    #	{"instanceId":"i-bp15wj5w33t8vf******"}
    #	]
    #	'''
`
)

type (
	Dimension struct {
		Name   string `toml:"name"`
		Value  string `toml:"value"`
		Period int    `toml:"period"`
	}

	Property struct {
		Name       string            `toml:"name"`
		Period     int               `toml:"period"`
		Interval   internal.Duration `toml:"interval"`
		Dimensions string            `toml:"dimensions"`
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
		Interval        internal.Duration `toml:"interval"`
		Delay           internal.Duration `toml:"delay"`
		Project         []*Project        `toml:"project"`
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
		q    *cms.DescribeMetricListRequest
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
					checkDimensions := []map[string]string{}
					if err := json.Unmarshal([]byte(prop.Dimensions), &checkDimensions); err != nil {
						return nil, fmt.Errorf("invalid dimension(%s): %s, %s", metric, prop.Dimensions, err)
					}
					p.globalMetricProperty = prop
					p.globalMetricProperty.Dimensions = strings.Trim(prop.Dimensions, " \t\r\n")
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
				break
			}
		}

		if property == nil && p.globalMetricProperty != nil {
			if p.globalMetricProperty.Period > 0 {
				req.Period = strconv.FormatInt(int64(p.globalMetricProperty.Period), 10)
			}
			interval = p.globalMetricProperty.Interval.Duration
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
