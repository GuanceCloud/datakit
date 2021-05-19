package v1

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
)

type CesClient struct {
	HcClient *http_client.HcHttpClient
}

func NewCesClient(hcClient *http_client.HcHttpClient) *CesClient {
	return &CesClient{HcClient: hcClient}
}

func CesClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//批量查询指定时间范围内指定指标的指定粒度的监控数据，目前最多支持10指标的批量查询。
func (c *CesClient) BatchListMetricData(request *model.BatchListMetricDataRequest) (*model.BatchListMetricDataResponse, error) {
	requestDef := GenReqDefForBatchListMetricData()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.BatchListMetricDataResponse), nil
	}
}

//创建一条告警规则。
func (c *CesClient) CreateAlarm(request *model.CreateAlarmRequest) (*model.CreateAlarmResponse, error) {
	requestDef := GenReqDefForCreateAlarm()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateAlarmResponse), nil
	}
}

//创建自定义告警模板。
func (c *CesClient) CreateAlarmTemplate(request *model.CreateAlarmTemplateRequest) (*model.CreateAlarmTemplateResponse, error) {
	requestDef := GenReqDefForCreateAlarmTemplate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateAlarmTemplateResponse), nil
	}
}

//上报自定义事件。
func (c *CesClient) CreateEvents(request *model.CreateEventsRequest) (*model.CreateEventsResponse, error) {
	requestDef := GenReqDefForCreateEvents()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateEventsResponse), nil
	}
}

//添加一条或多条指标监控数据。
func (c *CesClient) CreateMetricData(request *model.CreateMetricDataRequest) (*model.CreateMetricDataResponse, error) {
	requestDef := GenReqDefForCreateMetricData()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateMetricDataResponse), nil
	}
}

//创建资源分组，资源分组支持将各类资源按照业务集中进行分组管理，可以从分组角度查看监控与告警信息，以提升运维效率。
func (c *CesClient) CreateResourceGroup(request *model.CreateResourceGroupRequest) (*model.CreateResourceGroupResponse, error) {
	requestDef := GenReqDefForCreateResourceGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateResourceGroupResponse), nil
	}
}

//删除一条告警规则。
func (c *CesClient) DeleteAlarm(request *model.DeleteAlarmRequest) (*model.DeleteAlarmResponse, error) {
	requestDef := GenReqDefForDeleteAlarm()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteAlarmResponse), nil
	}
}

//根据ID删除自定义告警模板。
func (c *CesClient) DeleteAlarmTemplate(request *model.DeleteAlarmTemplateRequest) (*model.DeleteAlarmTemplateResponse, error) {
	requestDef := GenReqDefForDeleteAlarmTemplate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteAlarmTemplateResponse), nil
	}
}

//删除一条资源分组。
func (c *CesClient) DeleteResourceGroup(request *model.DeleteResourceGroupRequest) (*model.DeleteResourceGroupResponse, error) {
	requestDef := GenReqDefForDeleteResourceGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteResourceGroupResponse), nil
	}
}

//查询告警历史列表。
func (c *CesClient) ListAlarmHistories(request *model.ListAlarmHistoriesRequest) (*model.ListAlarmHistoriesResponse, error) {
	requestDef := GenReqDefForListAlarmHistories()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListAlarmHistoriesResponse), nil
	}
}

//查询自定义告警模板列表
func (c *CesClient) ListAlarmTemplates(request *model.ListAlarmTemplatesRequest) (*model.ListAlarmTemplatesResponse, error) {
	requestDef := GenReqDefForListAlarmTemplates()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListAlarmTemplatesResponse), nil
	}
}

//查询告警规则列表，可以指定分页条件限制结果数量，可以指定排序规则。
func (c *CesClient) ListAlarms(request *model.ListAlarmsRequest) (*model.ListAlarmsResponse, error) {
	requestDef := GenReqDefForListAlarms()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListAlarmsResponse), nil
	}
}

//根据事件监控名称，查询该事件发生的详细信息。
func (c *CesClient) ListEventDetail(request *model.ListEventDetailRequest) (*model.ListEventDetailResponse, error) {
	requestDef := GenReqDefForListEventDetail()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListEventDetailResponse), nil
	}
}

//查询事件监控列表，包括系统事件和自定义事件。
func (c *CesClient) ListEvents(request *model.ListEventsRequest) (*model.ListEventsResponse, error) {
	requestDef := GenReqDefForListEvents()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListEventsResponse), nil
	}
}

//查询系统当前可监控指标列表，可以指定指标命名空间、指标名称、维度、排序方式，起始记录和最大记录条数过滤查询结果。
func (c *CesClient) ListMetrics(request *model.ListMetricsRequest) (*model.ListMetricsResponse, error) {
	requestDef := GenReqDefForListMetrics()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListMetricsResponse), nil
	}
}

//查询所创建的所有资源分组。
func (c *CesClient) ListResourceGroup(request *model.ListResourceGroupRequest) (*model.ListResourceGroupResponse, error) {
	requestDef := GenReqDefForListResourceGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListResourceGroupResponse), nil
	}
}

//根据告警ID查询告警规则信息。
func (c *CesClient) ShowAlarm(request *model.ShowAlarmRequest) (*model.ShowAlarmResponse, error) {
	requestDef := GenReqDefForShowAlarm()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowAlarmResponse), nil
	}
}

//查询指定时间范围指定事件类型的主机配置数据，可以通过参数指定需要查询的数据维度。注意：该接口提供给HANA场景下SAP Monitor查询主机配置数据，其他场景下查不到主机配置数据。
func (c *CesClient) ShowEventData(request *model.ShowEventDataRequest) (*model.ShowEventDataResponse, error) {
	requestDef := GenReqDefForShowEventData()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowEventDataResponse), nil
	}
}

//查询指定时间范围指定指标的指定粒度的监控数据，可以通过参数指定需要查询的数据维度。
func (c *CesClient) ShowMetricData(request *model.ShowMetricDataRequest) (*model.ShowMetricDataResponse, error) {
	requestDef := GenReqDefForShowMetricData()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowMetricDataResponse), nil
	}
}

//查询用户可以创建的资源配额总数及当前使用量，当前仅有告警规则一种资源类型。
func (c *CesClient) ShowQuotas(request *model.ShowQuotasRequest) (*model.ShowQuotasResponse, error) {
	requestDef := GenReqDefForShowQuotas()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowQuotasResponse), nil
	}
}

//根据资源分组ID查询资源分组下的资源。
func (c *CesClient) ShowResourceGroup(request *model.ShowResourceGroupRequest) (*model.ShowResourceGroupResponse, error) {
	requestDef := GenReqDefForShowResourceGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowResourceGroupResponse), nil
	}
}

//修改告警规则。
func (c *CesClient) UpdateAlarm(request *model.UpdateAlarmRequest) (*model.UpdateAlarmResponse, error) {
	requestDef := GenReqDefForUpdateAlarm()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateAlarmResponse), nil
	}
}

//启动或停止一条告警规则。
func (c *CesClient) UpdateAlarmAction(request *model.UpdateAlarmActionRequest) (*model.UpdateAlarmActionResponse, error) {
	requestDef := GenReqDefForUpdateAlarmAction()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateAlarmActionResponse), nil
	}
}

//更新自定义告警模板。
func (c *CesClient) UpdateAlarmTemplate(request *model.UpdateAlarmTemplateRequest) (*model.UpdateAlarmTemplateResponse, error) {
	requestDef := GenReqDefForUpdateAlarmTemplate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateAlarmTemplateResponse), nil
	}
}

//更新资源分组，资源分组支持将各类资源按照业务集中进行分组管理，可以从分组角度查看监控与告警信息，以提升运维效率。
func (c *CesClient) UpdateResourceGroup(request *model.UpdateResourceGroupRequest) (*model.UpdateResourceGroupResponse, error) {
	requestDef := GenReqDefForUpdateResourceGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateResourceGroupResponse), nil
	}
}
