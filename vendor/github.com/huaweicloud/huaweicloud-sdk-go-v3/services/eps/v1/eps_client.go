package v1

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eps/v1/model"
)

type EpsClient struct {
	HcClient *http_client.HcHttpClient
}

func NewEpsClient(hcClient *http_client.HcHttpClient) *EpsClient {
	return &EpsClient{HcClient: hcClient}
}

func EpsClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder().WithCredentialsType("global.Credentials")
	return builder
}

//创建企业项目。
func (c *EpsClient) CreateEnterpriseProject(request *model.CreateEnterpriseProjectRequest) (*model.CreateEnterpriseProjectResponse, error) {
	requestDef := GenReqDefForCreateEnterpriseProject()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateEnterpriseProjectResponse), nil
	}
}

//停用企业项目。
func (c *EpsClient) DisableEnterpriseProject(request *model.DisableEnterpriseProjectRequest) (*model.DisableEnterpriseProjectResponse, error) {
	requestDef := GenReqDefForDisableEnterpriseProject()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DisableEnterpriseProjectResponse), nil
	}
}

//启用企业项目。
func (c *EpsClient) EnableEnterpriseProject(request *model.EnableEnterpriseProjectRequest) (*model.EnableEnterpriseProjectResponse, error) {
	requestDef := GenReqDefForEnableEnterpriseProject()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.EnableEnterpriseProjectResponse), nil
	}
}

//查询企业项目的API版本列表。
func (c *EpsClient) ListApiVersions(request *model.ListApiVersionsRequest) (*model.ListApiVersionsResponse, error) {
	requestDef := GenReqDefForListApiVersions()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListApiVersionsResponse), nil
	}
}

//查询当前用户已授权的企业项目列表，用户可以使用企业项目绑定资源。
func (c *EpsClient) ListEnterpriseProject(request *model.ListEnterpriseProjectRequest) (*model.ListEnterpriseProjectResponse, error) {
	requestDef := GenReqDefForListEnterpriseProject()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListEnterpriseProjectResponse), nil
	}
}

//迁移资源到目标企业项目。
func (c *EpsClient) MigrateResource(request *model.MigrateResourceRequest) (*model.MigrateResourceResponse, error) {
	requestDef := GenReqDefForMigrateResource()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.MigrateResourceResponse), nil
	}
}

//修改企业项目。当前仅支持修改名称和描述。
func (c *EpsClient) ModifyEnterpriseProject(request *model.ModifyEnterpriseProjectRequest) (*model.ModifyEnterpriseProjectResponse, error) {
	requestDef := GenReqDefForModifyEnterpriseProject()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ModifyEnterpriseProjectResponse), nil
	}
}

//查询指定的企业项目API版本号详情
func (c *EpsClient) ShowApiVersion(request *model.ShowApiVersionRequest) (*model.ShowApiVersionResponse, error) {
	requestDef := GenReqDefForShowApiVersion()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowApiVersionResponse), nil
	}
}

//查询企业项目详情。
func (c *EpsClient) ShowEnterpriseProject(request *model.ShowEnterpriseProjectRequest) (*model.ShowEnterpriseProjectResponse, error) {
	requestDef := GenReqDefForShowEnterpriseProject()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowEnterpriseProjectResponse), nil
	}
}

//查询企业项目的配额信息。
func (c *EpsClient) ShowEnterpriseProjectQuota(request *model.ShowEnterpriseProjectQuotaRequest) (*model.ShowEnterpriseProjectQuotaResponse, error) {
	requestDef := GenReqDefForShowEnterpriseProjectQuota()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowEnterpriseProjectQuotaResponse), nil
	}
}

//查询企业项目下绑定的资源详情。
func (c *EpsClient) ShowResourceBindEnterpriseProject(request *model.ShowResourceBindEnterpriseProjectRequest) (*model.ShowResourceBindEnterpriseProjectResponse, error) {
	requestDef := GenReqDefForShowResourceBindEnterpriseProject()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowResourceBindEnterpriseProjectResponse), nil
	}
}
