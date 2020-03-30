package ucmon

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
)

const (
	sampleConfig = `
#[[instances]]
#public_key = ''
#private_key = ''
#region = ''
#zone = ''
#project_id = ''

#[[instances.resource]]
#resource_type = ''
#resource_id = ''

#[[instances.resource.metrics]]
#metric_name = ''
# #interval = '5m' #default 5 minitues
`
)

type (
	ucMetric struct {
		MetricName string
		Interval   internal.Duration
	}

	ucResource struct {
		ResourceType string
		ResourceID   string
		Metrics      []*ucMetric
	}

	ucInstance struct {
		PublicKey  string
		PrivateKey string
		Region     string
		Zone       string
		ProjectID  string
		Resource   []*ucResource
	}

	metricMeta struct {
		unit string
	}

	queryListInfo struct {
		meta *metricMeta

		intervalTime time.Duration

		resourceID   string
		resourceType string
		metricname   string

		lastFetchTime time.Time
	}
)

func (a *ucInstance) genQueryInfo() []*queryListInfo {

	var infos []*queryListInfo

	for _, res := range a.Resource {
		for _, ms := range res.Metrics {
			info := &queryListInfo{
				resourceID:   res.ResourceID,
				resourceType: res.ResourceType,
				metricname:   ms.MetricName,
				intervalTime: ms.Interval.Duration,
			}

			if info.intervalTime == 0 {
				info.intervalTime = 5 * time.Minute
			} else if info.intervalTime < time.Minute {
				info.intervalTime = time.Minute
			}

			infos = append(infos, info)
		}
	}

	return infos
}
