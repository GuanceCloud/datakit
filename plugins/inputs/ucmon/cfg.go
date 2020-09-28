package ucmon

import (
	"context"
	"time"

	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	sampleConfig = `
#(required)
#[[inputs.ucloud_monitor]]

#(required)
#public_key = ''

#(required)
#private_key = ''

#(required)
#region = ''

#(required)
#zone = ''

#(optional) use the default project if empty. sub account must not be empty
#project_id = ''

#(required)
#[[inputs.ucloud_monitor.resource]]

# ##(required) resource type
#resource_type = ''

# ##(required) should be none-empty expect for 'sharebandwidth'(default use the first available resource id)
#resource_id = ''

# ##(optional)
#interval='5m'

# ##(required) names of metric to collect
#metrics=[]
`
)

type (
	ucResource struct {
		ResourceType string
		ResourceID   string
		Interval     datakit.Duration
		Metrics      []string
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

		debugMode bool
	}

	queryListInfo struct {
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
				metricname:   ms,
				intervalTime: res.Interval.Duration,
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
