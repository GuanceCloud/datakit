package v1

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/tms/v1/model"
)

type TmsClient struct {
	HcClient *http_client.HcHttpClient
}

func NewTmsClient(hcClient *http_client.HcHttpClient) *TmsClient {
	return &TmsClient{HcClient: hcClient}
}

func TmsClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//用于创建预定标签。用户创建预定义标签后，可以使用预定义标签来给资源创建标签。该接口支持幂等特性和处理批量数据。
func (c *TmsClient) CreatePredefineTags(request *model.CreatePredefineTagsRequest) (*model.CreatePredefineTagsResponse, error) {
	requestDef := GenReqDefForCreatePredefineTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreatePredefineTagsResponse), nil
	}
}

//用于删除预定标签。
func (c *TmsClient) DeletePredefineTags(request *model.DeletePredefineTagsRequest) (*model.DeletePredefineTagsResponse, error) {
	requestDef := GenReqDefForDeletePredefineTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeletePredefineTagsResponse), nil
	}
}

//查询标签管理服务的API版本列表。
func (c *TmsClient) ListApiVersions(request *model.ListApiVersionsRequest) (*model.ListApiVersionsResponse, error) {
	requestDef := GenReqDefForListApiVersions()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListApiVersionsResponse), nil
	}
}

//用于查询预定义标签列表。
func (c *TmsClient) ListPredefineTags(request *model.ListPredefineTagsRequest) (*model.ListPredefineTagsResponse, error) {
	requestDef := GenReqDefForListPredefineTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListPredefineTagsResponse), nil
	}
}

//查询指定的标签管理服务API版本号详情。
func (c *TmsClient) ShowApiVersion(request *model.ShowApiVersionRequest) (*model.ShowApiVersionResponse, error) {
	requestDef := GenReqDefForShowApiVersion()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowApiVersionResponse), nil
	}
}

//修改预定义标签。
func (c *TmsClient) UpdatePredefineTags(request *model.UpdatePredefineTagsRequest) (*model.UpdatePredefineTagsResponse, error) {
	requestDef := GenReqDefForUpdatePredefineTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdatePredefineTagsResponse), nil
	}
}
