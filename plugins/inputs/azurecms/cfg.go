package azurecms

import (
	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2019-06-01/insights"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
)

const (
	sampleConfig = `#[[instances]]
# client_id = ''
# client_secret = ''
# tenant_id = ''
# subscription_id = ''
# #end_point = 'https://management.chinacloudapi.cn'

#[[instances.resource]]
#resource_id = ''

#[[instances.resource.metrics]]
#metric_name = 'Percentage CPU'
#interval = '1m'
`
)

type (
	azureMetric struct {
		MetricName string
		Interval   internal.Duration
	}

	azureResource struct {
		ResourceID string
		Metrics    []*azureMetric
	}

	azureInstance struct {
		ClientID       string
		ClientSecret   string
		TenantID       string
		SubscriptionID string
		EndPoint       string //https://management.chinacloudapi.cn

		Resource []*azureResource
	}

	metricMeta struct {
		supportTimeGrain      []int64
		supportedAggregations []string
		unit                  string
	}

	queryListInfo struct {
		meta *metricMeta

		resourceID  string
		timespan    string
		interval    string
		metricname  string
		aggregation string
		top         int32
		orderby     string
		filter      string
		resultType  insights.ResultType
		//apiVersion  string // "2018-01-01"
		metricnamespace string // `Microsoft.Compute/virtualMachines`
	}
)

func (a *azureInstance) genQueryInfo() []*queryListInfo {

	var infos []*queryListInfo

	for _, res := range a.Resource {
		for _, ms := range res.Metrics {
			info := &queryListInfo{
				resourceID: res.ResourceID,
				metricname: ms.MetricName,
				interval:   "",
			}
			infos = append(infos, info)
		}
	}

	return infos
}
