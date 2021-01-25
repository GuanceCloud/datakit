package v2

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/servicestage/v2/model"
)

type ServiceStageClient struct {
	HcClient *http_client.HcHttpClient
}

func NewServiceStageClient(hcClient *http_client.HcHttpClient) *ServiceStageClient {
	return &ServiceStageClient{HcClient: hcClient}
}

func ServiceStageClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//此API通过应用ID修改应用信息。
func (c *ServiceStageClient) ChangeApplication(request *model.ChangeApplicationRequest) (*model.ChangeApplicationResponse, error) {
	requestDef := GenReqDefForChangeApplication()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ChangeApplicationResponse), nil
	}
}

//通过此API修改应用配置信息。
func (c *ServiceStageClient) ChangeApplicationConfiguration(request *model.ChangeApplicationConfigurationRequest) (*model.ChangeApplicationConfigurationResponse, error) {
	requestDef := GenReqDefForChangeApplicationConfiguration()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ChangeApplicationConfigurationResponse), nil
	}
}

//此API通过组件ID修改组件信息。
func (c *ServiceStageClient) ChangeComponent(request *model.ChangeComponentRequest) (*model.ChangeComponentResponse, error) {
	requestDef := GenReqDefForChangeComponent()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ChangeComponentResponse), nil
	}
}

//此API通过环境ID修改环境信息。
func (c *ServiceStageClient) ChangeEnvironment(request *model.ChangeEnvironmentRequest) (*model.ChangeEnvironmentResponse, error) {
	requestDef := GenReqDefForChangeEnvironment()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ChangeEnvironmentResponse), nil
	}
}

//通过此API修改应用组件实例。
func (c *ServiceStageClient) ChangeInstance(request *model.ChangeInstanceRequest) (*model.ChangeInstanceResponse, error) {
	requestDef := GenReqDefForChangeInstance()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ChangeInstanceResponse), nil
	}
}

//此API用来修改环境资源。
func (c *ServiceStageClient) ChangeResourceInEnvironment(request *model.ChangeResourceInEnvironmentRequest) (*model.ChangeResourceInEnvironmentResponse, error) {
	requestDef := GenReqDefForChangeResourceInEnvironment()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ChangeResourceInEnvironmentResponse), nil
	}
}

//应用是一个功能相对完备的业务系统，由一个或多个特性相关的组件组成。  此API用来创建应用。
func (c *ServiceStageClient) CreateApplication(request *model.CreateApplicationRequest) (*model.CreateApplicationResponse, error) {
	requestDef := GenReqDefForCreateApplication()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateApplicationResponse), nil
	}
}

//应用组件是组成应用的某个业务特性实现，以代码或者软件包为载体，可独立部署运行。  此API用来在应用中创建组件。
func (c *ServiceStageClient) CreateComponent(request *model.CreateComponentRequest) (*model.CreateComponentResponse, error) {
	requestDef := GenReqDefForCreateComponent()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateComponentResponse), nil
	}
}

//环境是用于应用部署和运行的计算、存储、网络等基础设施的集合。Servicestage把相同VPC下的CCE集群加上多个ELB、RDS、DCS实例组合为一个环境，如：开发环境，测试环境，预生产环境，生产环境。环境内网络互通，可以按环境维度来管理资源、部署服务，减少具体基础设施运维管理的复杂性。  此API用来创建环境。
func (c *ServiceStageClient) CreateEnvironment(request *model.CreateEnvironmentRequest) (*model.CreateEnvironmentResponse, error) {
	requestDef := GenReqDefForCreateEnvironment()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateEnvironmentResponse), nil
	}
}

//此API用来创建应用组件实例。
func (c *ServiceStageClient) CreateInstance(request *model.CreateInstanceRequest) (*model.CreateInstanceResponse, error) {
	requestDef := GenReqDefForCreateInstance()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateInstanceResponse), nil
	}
}

//此API通过应用ID删除应用。
func (c *ServiceStageClient) DeleteApplication(request *model.DeleteApplicationRequest) (*model.DeleteApplicationResponse, error) {
	requestDef := GenReqDefForDeleteApplication()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteApplicationResponse), nil
	}
}

//通过此API删除应用配置信息。
func (c *ServiceStageClient) DeleteApplicationConfiguration(request *model.DeleteApplicationConfigurationRequest) (*model.DeleteApplicationConfigurationResponse, error) {
	requestDef := GenReqDefForDeleteApplicationConfiguration()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteApplicationConfigurationResponse), nil
	}
}

//此API通过应用组件ID删除应用组件。
func (c *ServiceStageClient) DeleteComponent(request *model.DeleteComponentRequest) (*model.DeleteComponentResponse, error) {
	requestDef := GenReqDefForDeleteComponent()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteComponentResponse), nil
	}
}

//此API通过环境ID删除环境。
func (c *ServiceStageClient) DeleteEnvironment(request *model.DeleteEnvironmentRequest) (*model.DeleteEnvironmentResponse, error) {
	requestDef := GenReqDefForDeleteEnvironment()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteEnvironmentResponse), nil
	}
}

//通过此API删除应用组件实例。
func (c *ServiceStageClient) DeleteInstance(request *model.DeleteInstanceRequest) (*model.DeleteInstanceResponse, error) {
	requestDef := GenReqDefForDeleteInstance()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteInstanceResponse), nil
	}
}

//通过此API可以获取所有已经创建的应用。
func (c *ServiceStageClient) ListApplications(request *model.ListApplicationsRequest) (*model.ListApplicationsResponse, error) {
	requestDef := GenReqDefForListApplications()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListApplicationsResponse), nil
	}
}

//通过此API获取应用下所有应用组件。
func (c *ServiceStageClient) ListComponents(request *model.ListComponentsRequest) (*model.ListComponentsResponse, error) {
	requestDef := GenReqDefForListComponents()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListComponentsResponse), nil
	}
}

//此API用来获取所有已经创建环境。
func (c *ServiceStageClient) ListEnvironments(request *model.ListEnvironmentsRequest) (*model.ListEnvironmentsResponse, error) {
	requestDef := GenReqDefForListEnvironments()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListEnvironmentsResponse), nil
	}
}

//通过此API获取应用组件实例快照信息。
func (c *ServiceStageClient) ListInstanceSnapshots(request *model.ListInstanceSnapshotsRequest) (*model.ListInstanceSnapshotsResponse, error) {
	requestDef := GenReqDefForListInstanceSnapshots()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListInstanceSnapshotsResponse), nil
	}
}

//通过此API获取组件下的所有组件实例。
func (c *ServiceStageClient) ListInstances(request *model.ListInstancesRequest) (*model.ListInstancesResponse, error) {
	requestDef := GenReqDefForListInstances()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListInstancesResponse), nil
	}
}

//通过此API获取应用配置信息。
func (c *ServiceStageClient) ShowApplicationConfiguration(request *model.ShowApplicationConfigurationRequest) (*model.ShowApplicationConfigurationResponse, error) {
	requestDef := GenReqDefForShowApplicationConfiguration()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowApplicationConfigurationResponse), nil
	}
}

//此API通过应用ID获取应用详细信息。
func (c *ServiceStageClient) ShowApplicationDetail(request *model.ShowApplicationDetailRequest) (*model.ShowApplicationDetailResponse, error) {
	requestDef := GenReqDefForShowApplicationDetail()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowApplicationDetailResponse), nil
	}
}

//通过组件ID获取应用组件信息。
func (c *ServiceStageClient) ShowComponentDetail(request *model.ShowComponentDetailRequest) (*model.ShowComponentDetailResponse, error) {
	requestDef := GenReqDefForShowComponentDetail()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowComponentDetailResponse), nil
	}
}

//此API通过环境ID获取环境详细信息。
func (c *ServiceStageClient) ShowEnvironmentDetail(request *model.ShowEnvironmentDetailRequest) (*model.ShowEnvironmentDetailResponse, error) {
	requestDef := GenReqDefForShowEnvironmentDetail()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowEnvironmentDetailResponse), nil
	}
}

//此API通过实例ID获取实例详细信息。
func (c *ServiceStageClient) ShowInstanceDetail(request *model.ShowInstanceDetailRequest) (*model.ShowInstanceDetailResponse, error) {
	requestDef := GenReqDefForShowInstanceDetail()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowInstanceDetailResponse), nil
	}
}

//通过此API获取部署任务详细信息。
func (c *ServiceStageClient) ShowJobDetail(request *model.ShowJobDetailRequest) (*model.ShowJobDetailResponse, error) {
	requestDef := GenReqDefForShowJobDetail()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowJobDetailResponse), nil
	}
}

//通过此API获取对组件实例的操作。
func (c *ServiceStageClient) UpdateInstanceAction(request *model.UpdateInstanceActionRequest) (*model.UpdateInstanceActionResponse, error) {
	requestDef := GenReqDefForUpdateInstanceAction()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateInstanceActionResponse), nil
	}
}

//在指定仓库项目下创建文件。
func (c *ServiceStageClient) CreateFile(request *model.CreateFileRequest) (*model.CreateFileResponse, error) {
	requestDef := GenReqDefForCreateFile()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateFileResponse), nil
	}
}

//创建指定项目的hook。
func (c *ServiceStageClient) CreateHook(request *model.CreateHookRequest) (*model.CreateHookResponse, error) {
	requestDef := GenReqDefForCreateHook()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateHookResponse), nil
	}
}

//创建指定Git仓库类型的OAuth授权。
func (c *ServiceStageClient) CreateOAuth(request *model.CreateOAuthRequest) (*model.CreateOAuthResponse, error) {
	requestDef := GenReqDefForCreateOAuth()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateOAuthResponse), nil
	}
}

//创建指定Git仓库类型的口令授权。
func (c *ServiceStageClient) CreatePasswordAuth(request *model.CreatePasswordAuthRequest) (*model.CreatePasswordAuthResponse, error) {
	requestDef := GenReqDefForCreatePasswordAuth()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreatePasswordAuthResponse), nil
	}
}

//创建指定Git仓库类型的私人令牌授权。
func (c *ServiceStageClient) CreatePersonalAuth(request *model.CreatePersonalAuthRequest) (*model.CreatePersonalAuthResponse, error) {
	requestDef := GenReqDefForCreatePersonalAuth()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreatePersonalAuthResponse), nil
	}
}

//创建指定组织下的软件仓库项目。
func (c *ServiceStageClient) CreateProject(request *model.CreateProjectRequest) (*model.CreateProjectResponse, error) {
	requestDef := GenReqDefForCreateProject()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateProjectResponse), nil
	}
}

//创建指定项目的tag标签。
func (c *ServiceStageClient) CreateTag(request *model.CreateTagRequest) (*model.CreateTagResponse, error) {
	requestDef := GenReqDefForCreateTag()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateTagResponse), nil
	}
}

//通过名称删除仓库授权。
func (c *ServiceStageClient) DeleteAuthorize(request *model.DeleteAuthorizeRequest) (*model.DeleteAuthorizeResponse, error) {
	requestDef := GenReqDefForDeleteAuthorize()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteAuthorizeResponse), nil
	}
}

//删除指定项目仓库下的文件。
func (c *ServiceStageClient) DeleteFile(request *model.DeleteFileRequest) (*model.DeleteFileResponse, error) {
	requestDef := GenReqDefForDeleteFile()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteFileResponse), nil
	}
}

//删除指定项目的hook。
func (c *ServiceStageClient) DeleteHook(request *model.DeleteHookRequest) (*model.DeleteHookResponse, error) {
	requestDef := GenReqDefForDeleteHook()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteHookResponse), nil
	}
}

//删除指定项目的tag标签。
func (c *ServiceStageClient) DeleteTag(request *model.DeleteTagRequest) (*model.DeleteTagResponse, error) {
	requestDef := GenReqDefForDeleteTag()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteTagResponse), nil
	}
}

//获取所有Git仓库授权信息。
func (c *ServiceStageClient) ListAuthorizations(request *model.ListAuthorizationsRequest) (*model.ListAuthorizationsResponse, error) {
	requestDef := GenReqDefForListAuthorizations()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListAuthorizationsResponse), nil
	}
}

//获取指定项目的所有分支列表。
func (c *ServiceStageClient) ListBranches(request *model.ListBranchesRequest) (*model.ListBranchesResponse, error) {
	requestDef := GenReqDefForListBranches()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListBranchesResponse), nil
	}
}

//获取指定项目的最近10次commit提交记录。
func (c *ServiceStageClient) ListCommits(request *model.ListCommitsRequest) (*model.ListCommitsResponse, error) {
	requestDef := GenReqDefForListCommits()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListCommitsResponse), nil
	}
}

//获取指定项目的所有hooks
func (c *ServiceStageClient) ListHooks(request *model.ListHooksRequest) (*model.ListHooksResponse, error) {
	requestDef := GenReqDefForListHooks()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListHooksResponse), nil
	}
}

//获取仓库的namespaces。
func (c *ServiceStageClient) ListNamespaces(request *model.ListNamespacesRequest) (*model.ListNamespacesResponse, error) {
	requestDef := GenReqDefForListNamespaces()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListNamespacesResponse), nil
	}
}

//获取指定组织下的所有项目。
func (c *ServiceStageClient) ListProjects(request *model.ListProjectsRequest) (*model.ListProjectsResponse, error) {
	requestDef := GenReqDefForListProjects()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListProjectsResponse), nil
	}
}

//获取指定项目的所有tag标签。
func (c *ServiceStageClient) ListTags(request *model.ListTagsRequest) (*model.ListTagsResponse, error) {
	requestDef := GenReqDefForListTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListTagsResponse), nil
	}
}

//获取指定项目仓库的文件列表。
func (c *ServiceStageClient) ListTrees(request *model.ListTreesRequest) (*model.ListTreesResponse, error) {
	requestDef := GenReqDefForListTrees()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListTreesResponse), nil
	}
}

//获取指定项目仓库下文件的内容。
func (c *ServiceStageClient) ShowContent(request *model.ShowContentRequest) (*model.ShowContentResponse, error) {
	requestDef := GenReqDefForShowContent()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowContentResponse), nil
	}
}

//通过指定的clone url 获取仓库信息。
func (c *ServiceStageClient) ShowProjectDetail(request *model.ShowProjectDetailRequest) (*model.ShowProjectDetailResponse, error) {
	requestDef := GenReqDefForShowProjectDetail()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowProjectDetailResponse), nil
	}
}

//获取指定Git仓库类型的授权重定向URL。
func (c *ServiceStageClient) ShowRedirectUrl(request *model.ShowRedirectUrlRequest) (*model.ShowRedirectUrlResponse, error) {
	requestDef := GenReqDefForShowRedirectUrl()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowRedirectUrlResponse), nil
	}
}

//更新指定项目仓库下的文件内容。
func (c *ServiceStageClient) UpdateFile(request *model.UpdateFileRequest) (*model.UpdateFileResponse, error) {
	requestDef := GenReqDefForUpdateFile()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateFileResponse), nil
	}
}

//通过此API获取所用支持的应用资源规格。
func (c *ServiceStageClient) ListFlavors(request *model.ListFlavorsRequest) (*model.ListFlavorsResponse, error) {
	requestDef := GenReqDefForListFlavors()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListFlavorsResponse), nil
	}
}

//此API用来获取所有支持应用组件运行时类型。
func (c *ServiceStageClient) ListRuntimes(request *model.ListRuntimesRequest) (*model.ListRuntimesResponse, error) {
	requestDef := GenReqDefForListRuntimes()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListRuntimesResponse), nil
	}
}

//此API用来获取所有内置应用组件模板。
func (c *ServiceStageClient) ListTemplates(request *model.ListTemplatesRequest) (*model.ListTemplatesResponse, error) {
	requestDef := GenReqDefForListTemplates()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListTemplatesResponse), nil
	}
}
