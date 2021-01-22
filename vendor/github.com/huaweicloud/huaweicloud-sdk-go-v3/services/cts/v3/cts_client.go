package v3

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cts/v3/model"
)

type CtsClient struct {
	HcClient *http_client.HcHttpClient
}

func NewCtsClient(hcClient *http_client.HcHttpClient) *CtsClient {
	return &CtsClient{HcClient: hcClient}
}

func CtsClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//云审计服务开通后系统会自动创建一个追踪器，用来关联系统记录的所有操作。目前，一个云账户在一个Region下支持创建一个管理类追踪器和多个数据类追踪器。 云审计服务支持在管理控制台查询近7天内的操作记录。如需保存更长时间的操作记录，您可以在创建追踪器之后通过对象存储服务（Object Storage Service，以下简称OBS）将操作记录实时保存至OBS桶中。
func (c *CtsClient) CreateTracker(request *model.CreateTrackerRequest) (*model.CreateTrackerResponse, error) {
	requestDef := GenReqDefForCreateTracker()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateTrackerResponse), nil
	}
}

//云审计服务目前仅支持删除已创建的数据类追踪器。删除追踪器对已有的操作记录没有影响，当您重新开通云审计服务后，依旧可以查看已有的操作记录。
func (c *CtsClient) DeleteTracker(request *model.DeleteTrackerRequest) (*model.DeleteTrackerResponse, error) {
	requestDef := GenReqDefForDeleteTracker()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteTrackerResponse), nil
	}
}

//查询租户追踪器配额信息。
func (c *CtsClient) ListQuotas(request *model.ListQuotasRequest) (*model.ListQuotasResponse, error) {
	requestDef := GenReqDefForListQuotas()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListQuotasResponse), nil
	}
}

//通过事件列表查询接口，可以查出系统记录的7天内资源操作记录。
func (c *CtsClient) ListTraces(request *model.ListTracesRequest) (*model.ListTracesResponse, error) {
	requestDef := GenReqDefForListTraces()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListTracesResponse), nil
	}
}

//开通云审计服务成功后，您可以在追踪器信息页面查看追踪器的详细信息。详细信息主要包括追踪器名称，用于存储操作事件的OBS桶名称和OBS桶中的事件文件前缀。
func (c *CtsClient) ListTrackers(request *model.ListTrackersRequest) (*model.ListTrackersResponse, error) {
	requestDef := GenReqDefForListTrackers()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListTrackersResponse), nil
	}
}

//云审计服务支持修改已创建追踪器的配置项，包括OBS桶转储、关键事件通知、事件转储加密、通过LTS对管理类事件进行检索、事件文件完整性校验以及追踪器启停状态等相关参数，修改追踪器对已有的操作记录没有影响。修改追踪器完成后，系统立即以新的规则开始记录操作。
func (c *CtsClient) UpdateTracker(request *model.UpdateTrackerRequest) (*model.UpdateTrackerResponse, error) {
	requestDef := GenReqDefForUpdateTracker()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateTrackerResponse), nil
	}
}
