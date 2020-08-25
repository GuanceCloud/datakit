package ucmon

import (
	"context"
	"time"

	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	sampleConfig = `
#[[inputs.ucloud_monitor]]
#public_key = ''
#private_key = ''
#region = ''
#zone = ''
#project_id = ''

#[[inputs.ucloud_monitor.resource]]
#resource_type = ''
#resource_id = ''

#[[inputs.ucloud_monitor.resource.metrics]]
#metric_name = ''
# #interval = '5m' #default 5 minitues
`
)

type (
	ucMetric struct {
		MetricName string
		Interval   datakit.Duration
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

		queryInfos []*queryListInfo

		ucCli *ucloud.Client

		ctx       context.Context
		cancelFun context.CancelFunc
	}

	queryListInfo struct {
		//meta *metricMeta

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
