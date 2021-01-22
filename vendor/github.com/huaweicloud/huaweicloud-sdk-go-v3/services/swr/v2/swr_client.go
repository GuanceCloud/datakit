package v2

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/swr/v2/model"
)

type SwrClient struct {
	HcClient *http_client.HcHttpClient
}

func NewSwrClient(hcClient *http_client.HcHttpClient) *SwrClient {
	return &SwrClient{HcClient: hcClient}
}

func SwrClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//创建镜像自动同步任务
func (c *SwrClient) CreateImageSyncRepo(request *model.CreateImageSyncRepoRequest) (*model.CreateImageSyncRepoResponse, error) {
	requestDef := GenReqDefForCreateImageSyncRepo()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateImageSyncRepoResponse), nil
	}
}

//创建组织
func (c *SwrClient) CreateNamespace(request *model.CreateNamespaceRequest) (*model.CreateNamespaceResponse, error) {
	requestDef := GenReqDefForCreateNamespace()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateNamespaceResponse), nil
	}
}

//创建组织权限
func (c *SwrClient) CreateNamespaceAuth(request *model.CreateNamespaceAuthRequest) (*model.CreateNamespaceAuthResponse, error) {
	requestDef := GenReqDefForCreateNamespaceAuth()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateNamespaceAuthResponse), nil
	}
}

//创建共享账号。镜像上传后，您可以共享私有镜像给其他帐号，并授予下载该镜像的权限。
func (c *SwrClient) CreateRepoDomains(request *model.CreateRepoDomainsRequest) (*model.CreateRepoDomainsResponse, error) {
	requestDef := GenReqDefForCreateRepoDomains()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateRepoDomainsResponse), nil
	}
}

//创建镜像老化规则
func (c *SwrClient) CreateRetention(request *model.CreateRetentionRequest) (*model.CreateRetentionResponse, error) {
	requestDef := GenReqDefForCreateRetention()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateRetentionResponse), nil
	}
}

//调用该接口，通过获取响应消息头的X-Swr-Dockerlogin的值及响应消息体的host值，可生成临时登录指令。
func (c *SwrClient) CreateSecret(request *model.CreateSecretRequest) (*model.CreateSecretResponse, error) {
	requestDef := GenReqDefForCreateSecret()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateSecretResponse), nil
	}
}

//创建触发器
func (c *SwrClient) CreateTrigger(request *model.CreateTriggerRequest) (*model.CreateTriggerResponse, error) {
	requestDef := GenReqDefForCreateTrigger()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateTriggerResponse), nil
	}
}

//创建镜像权限
func (c *SwrClient) CreateUserRepositoryAuth(request *model.CreateUserRepositoryAuthRequest) (*model.CreateUserRepositoryAuthResponse, error) {
	requestDef := GenReqDefForCreateUserRepositoryAuth()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateUserRepositoryAuthResponse), nil
	}
}

//删除镜像自动同步任务
func (c *SwrClient) DeleteImageSyncRepo(request *model.DeleteImageSyncRepoRequest) (*model.DeleteImageSyncRepoResponse, error) {
	requestDef := GenReqDefForDeleteImageSyncRepo()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteImageSyncRepoResponse), nil
	}
}

//删除组织权限
func (c *SwrClient) DeleteNamespaceAuth(request *model.DeleteNamespaceAuthRequest) (*model.DeleteNamespaceAuthResponse, error) {
	requestDef := GenReqDefForDeleteNamespaceAuth()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteNamespaceAuthResponse), nil
	}
}

//删除组织
func (c *SwrClient) DeleteNamespaces(request *model.DeleteNamespacesRequest) (*model.DeleteNamespacesResponse, error) {
	requestDef := GenReqDefForDeleteNamespaces()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteNamespacesResponse), nil
	}
}

//删除组织下的镜像仓库。
func (c *SwrClient) DeleteRepo(request *model.DeleteRepoRequest) (*model.DeleteRepoResponse, error) {
	requestDef := GenReqDefForDeleteRepo()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteRepoResponse), nil
	}
}

//删除共享账号
func (c *SwrClient) DeleteRepoDomains(request *model.DeleteRepoDomainsRequest) (*model.DeleteRepoDomainsResponse, error) {
	requestDef := GenReqDefForDeleteRepoDomains()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteRepoDomainsResponse), nil
	}
}

//删除镜像仓库中指定tag的镜像
func (c *SwrClient) DeleteRepoTag(request *model.DeleteRepoTagRequest) (*model.DeleteRepoTagResponse, error) {
	requestDef := GenReqDefForDeleteRepoTag()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteRepoTagResponse), nil
	}
}

//删除镜像老化规则
func (c *SwrClient) DeleteRetention(request *model.DeleteRetentionRequest) (*model.DeleteRetentionResponse, error) {
	requestDef := GenReqDefForDeleteRetention()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteRetentionResponse), nil
	}
}

//删除触发器
func (c *SwrClient) DeleteTrigger(request *model.DeleteTriggerRequest) (*model.DeleteTriggerResponse, error) {
	requestDef := GenReqDefForDeleteTrigger()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteTriggerResponse), nil
	}
}

//删除镜像权限
func (c *SwrClient) DeleteUserRepositoryAuth(request *model.DeleteUserRepositoryAuthRequest) (*model.DeleteUserRepositoryAuthResponse, error) {
	requestDef := GenReqDefForDeleteUserRepositoryAuth()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteUserRepositoryAuthResponse), nil
	}
}

//获取镜像自动同步任务列表
func (c *SwrClient) ListImageAutoSyncReposDetails(request *model.ListImageAutoSyncReposDetailsRequest) (*model.ListImageAutoSyncReposDetailsResponse, error) {
	requestDef := GenReqDefForListImageAutoSyncReposDetails()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListImageAutoSyncReposDetailsResponse), nil
	}
}

//查询组织列表
func (c *SwrClient) ListNamespaces(request *model.ListNamespacesRequest) (*model.ListNamespacesResponse, error) {
	requestDef := GenReqDefForListNamespaces()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListNamespacesResponse), nil
	}
}

//获取共享账号列表
func (c *SwrClient) ListRepoDomains(request *model.ListRepoDomainsRequest) (*model.ListRepoDomainsResponse, error) {
	requestDef := GenReqDefForListRepoDomains()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListRepoDomainsResponse), nil
	}
}

//查询镜像tag列表
func (c *SwrClient) ListRepositoryTags(request *model.ListRepositoryTagsRequest) (*model.ListRepositoryTagsResponse, error) {
	requestDef := GenReqDefForListRepositoryTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListRepositoryTagsResponse), nil
	}
}

//获取镜像老化记录
func (c *SwrClient) ListRetentionHistories(request *model.ListRetentionHistoriesRequest) (*model.ListRetentionHistoriesResponse, error) {
	requestDef := GenReqDefForListRetentionHistories()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListRetentionHistoriesResponse), nil
	}
}

//获取镜像老化规则列表
func (c *SwrClient) ListRetentions(request *model.ListRetentionsRequest) (*model.ListRetentionsResponse, error) {
	requestDef := GenReqDefForListRetentions()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListRetentionsResponse), nil
	}
}

//获取镜像仓库下的触发器列表
func (c *SwrClient) ListTriggersDetails(request *model.ListTriggersDetailsRequest) (*model.ListTriggersDetailsResponse, error) {
	requestDef := GenReqDefForListTriggersDetails()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListTriggersDetailsResponse), nil
	}
}

//判断共享租户是否存在
func (c *SwrClient) ShowAccessDomain(request *model.ShowAccessDomainRequest) (*model.ShowAccessDomainResponse, error) {
	requestDef := GenReqDefForShowAccessDomain()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowAccessDomainResponse), nil
	}
}

//获取组织详情
func (c *SwrClient) ShowNamespace(request *model.ShowNamespaceRequest) (*model.ShowNamespaceResponse, error) {
	requestDef := GenReqDefForShowNamespace()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowNamespaceResponse), nil
	}
}

//查询组织权限
func (c *SwrClient) ShowNamespaceAuth(request *model.ShowNamespaceAuthRequest) (*model.ShowNamespaceAuthResponse, error) {
	requestDef := GenReqDefForShowNamespaceAuth()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowNamespaceAuthResponse), nil
	}
}

//查询镜像概要信息
func (c *SwrClient) ShowRepository(request *model.ShowRepositoryRequest) (*model.ShowRepositoryResponse, error) {
	requestDef := GenReqDefForShowRepository()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowRepositoryResponse), nil
	}
}

//获取镜像老化规则记录
func (c *SwrClient) ShowRetention(request *model.ShowRetentionRequest) (*model.ShowRetentionResponse, error) {
	requestDef := GenReqDefForShowRetention()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowRetentionResponse), nil
	}
}

//获取触发器详情
func (c *SwrClient) ShowTrigger(request *model.ShowTriggerRequest) (*model.ShowTriggerResponse, error) {
	requestDef := GenReqDefForShowTrigger()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowTriggerResponse), nil
	}
}

//查询镜像权限
func (c *SwrClient) ShowUserRepositoryAuth(request *model.ShowUserRepositoryAuthRequest) (*model.ShowUserRepositoryAuthResponse, error) {
	requestDef := GenReqDefForShowUserRepositoryAuth()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowUserRepositoryAuthResponse), nil
	}
}

//更新共享账号
func (c *SwrClient) UpdateRepoDomains(request *model.UpdateRepoDomainsRequest) (*model.UpdateRepoDomainsResponse, error) {
	requestDef := GenReqDefForUpdateRepoDomains()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateRepoDomainsResponse), nil
	}
}

//修改镜像老化规则
func (c *SwrClient) UpdateRetention(request *model.UpdateRetentionRequest) (*model.UpdateRetentionResponse, error) {
	requestDef := GenReqDefForUpdateRetention()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateRetentionResponse), nil
	}
}

//更新触发器配置
func (c *SwrClient) UpdateTrigger(request *model.UpdateTriggerRequest) (*model.UpdateTriggerResponse, error) {
	requestDef := GenReqDefForUpdateTrigger()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateTriggerResponse), nil
	}
}

//更新镜像权限
func (c *SwrClient) UpdateUserRepositoryAuth(request *model.UpdateUserRepositoryAuthRequest) (*model.UpdateUserRepositoryAuthResponse, error) {
	requestDef := GenReqDefForUpdateUserRepositoryAuth()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateUserRepositoryAuthResponse), nil
	}
}

//查询指定API版本信息
func (c *SwrClient) ShowApiVersion(request *model.ShowApiVersionRequest) (*model.ShowApiVersionResponse, error) {
	requestDef := GenReqDefForShowApiVersion()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowApiVersionResponse), nil
	}
}
