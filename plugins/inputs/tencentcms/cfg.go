package tencentcms

import (
	"encoding/json"
	"fmt"

	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
)

type (
	Dimension struct {
		Name  string `toml:"name"`
		Value string `toml:"value"`
	}

	Metrics struct {
		MetricNames []string     `toml:"names"`
		Dimensions  []*Dimension `toml:"dimensions"`
	}

	Namespace struct {
		Name    string   `toml:"name"`
		Metrics *Metrics `toml:"metrics"`
	}

	CMS struct {
		AccessKeyID     string       `toml:"access_key_id"`
		AccessKeySecret string       `toml:"access_key_secret"`
		RegionID        string       `toml:"region_id"`
		Namespace       []*Namespace `toml:"namespace"`
	}

	MetricsRequest struct {
		q           *monitor.GetMonitorDataRequest
		checkPeriod bool
	}
)

var (
	//Cfg             CMSConfig
	MetricsRequests = []*MetricsRequest{}
)

const (
	cmsConfigSample = `
#[[cms]]
#access_key_id = ""
#access_key_secret = ""

# ##See: https://cloud.tencent.com/document/product/213/6091
#region_id = 'ap-shanghai'


#[[cms.namespace]]
#	name='QCE/CVM'

#   ## Metrics to Pull (Required), See: https://cloud.tencent.com/document/api/248/30384
#	[cms.namespace.metrics]
#	names = [
#		"CPUUsage",
#	]

#     ## dimensions can be used to query the specified resource, which is a collection of key-value forms. 
#     ## each metric may have its own dimensions, See: https://cloud.tencent.com/document/api/248/30384
#     ## name is metric name, value is json
#	[[cms.namespace.metrics.dimensions]]
#		name = "CPUUsage"
#		value = '''
#		[
#			{"Dimensions":
#			[
#				{ "Name": "InstanceId", "Value": "ins-9bpjauir" }
#			]
#			}
#		]'''
`
)

func (p *Namespace) MakeDimension(mestric string) ([]*monitor.Instance, error) {

	var dimension *Dimension
	for _, d := range p.Metrics.Dimensions {
		if d.Name == mestric {
			dimension = d
			break
		}
	}

	if dimension == nil {
		return nil, nil
	}

	var insts []*monitor.Instance

	if dimension.Value != "" {
		if err := json.Unmarshal([]byte(dimension.Value), &insts); err != nil {
			return nil, fmt.Errorf("Dimension config of %s is invalid: %s", mestric, err)
		}
	}

	return insts, nil

}
