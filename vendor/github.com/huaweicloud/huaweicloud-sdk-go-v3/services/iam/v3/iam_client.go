package v3

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
)

type IamClient struct {
	HcClient *http_client.HcHttpClient
}

func NewIamClient(hcClient *http_client.HcHttpClient) *IamClient {
	return &IamClient{HcClient: hcClient}
}

func IamClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder().WithCredentialsType("global.Credentials,basic.Credentials")
	return builder
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)为委托授予所有项目服务权限。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) AssociateAgencyWithAllProjectsPermission(request *model.AssociateAgencyWithAllProjectsPermissionRequest) (*model.AssociateAgencyWithAllProjectsPermissionResponse, error) {
	requestDef := GenReqDefForAssociateAgencyWithAllProjectsPermission()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.AssociateAgencyWithAllProjectsPermissionResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)为委托授予全局服务权限。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) AssociateAgencyWithDomainPermission(request *model.AssociateAgencyWithDomainPermissionRequest) (*model.AssociateAgencyWithDomainPermissionResponse, error) {
	requestDef := GenReqDefForAssociateAgencyWithDomainPermission()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.AssociateAgencyWithDomainPermissionResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)为委托授予项目服务权限。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) AssociateAgencyWithProjectPermission(request *model.AssociateAgencyWithProjectPermissionRequest) (*model.AssociateAgencyWithProjectPermissionResponse, error) {
	requestDef := GenReqDefForAssociateAgencyWithProjectPermission()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.AssociateAgencyWithProjectPermissionResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)检查委托是否具有所有项目服务权限。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) CheckAllProjectsPermissionForAgency(request *model.CheckAllProjectsPermissionForAgencyRequest) (*model.CheckAllProjectsPermissionForAgencyResponse, error) {
	requestDef := GenReqDefForCheckAllProjectsPermissionForAgency()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CheckAllProjectsPermissionForAgencyResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询委托是否拥有全局服务权限。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) CheckDomainPermissionForAgency(request *model.CheckDomainPermissionForAgencyRequest) (*model.CheckDomainPermissionForAgencyResponse, error) {
	requestDef := GenReqDefForCheckDomainPermissionForAgency()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CheckDomainPermissionForAgencyResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询委托是否拥有项目服务权限。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) CheckProjectPermissionForAgency(request *model.CheckProjectPermissionForAgencyRequest) (*model.CheckProjectPermissionForAgencyResponse, error) {
	requestDef := GenReqDefForCheckProjectPermissionForAgency()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CheckProjectPermissionForAgencyResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)创建委托。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) CreateAgency(request *model.CreateAgencyRequest) (*model.CreateAgencyResponse, error) {
	requestDef := GenReqDefForCreateAgency()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateAgencyResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)创建委托自定义策略。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) CreateAgencyCustomPolicy(request *model.CreateAgencyCustomPolicyRequest) (*model.CreateAgencyCustomPolicyResponse, error) {
	requestDef := GenReqDefForCreateAgencyCustomPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateAgencyCustomPolicyResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)创建云服务自定义策略。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) CreateCloudServiceCustomPolicy(request *model.CreateCloudServiceCustomPolicyRequest) (*model.CreateCloudServiceCustomPolicyResponse, error) {
	requestDef := GenReqDefForCreateCloudServiceCustomPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateCloudServiceCustomPolicyResponse), nil
	}
}

//该接口用于用于获取自定义代理登录票据logintoken。logintoken是系统颁发给自定义代理用户的登录票据，承载用户的身份、session等信息。调用自定义代理URL登录云服务控制台时，可以使用本接口获取的logintoken进行认证。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。    > - logintoken的有效期为10分钟。
func (c *IamClient) CreateLoginToken(request *model.CreateLoginTokenRequest) (*model.CreateLoginTokenResponse, error) {
	requestDef := GenReqDefForCreateLoginToken()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateLoginTokenResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)导入Metadata文件。    账号在使用联邦认证功能前，需要先将Metadata文件导入到IAM中。Metadata文件是SAML 2.0协议约定的接口文件，包含访问接口地址和证书信息，请找企业管理员获取企业IdP的Metadata文件。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) CreateMetadata(request *model.CreateMetadataRequest) (*model.CreateMetadataResponse, error) {
	requestDef := GenReqDefForCreateMetadata()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateMetadataResponse), nil
	}
}

//该接口可以用于通过IdP initiated的联邦认证方式获取unscoped token。    Unscoped token不能用来鉴权，若联邦用户需要使用token进行鉴权，请参考[获取联邦认证scoped token](https://support.huaweicloud.com/api-iam/iam_13_0604.html)获取scoped token。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。     > - 该接口支持在命令行侧调用，需要客户端使用IdP initiated的联邦认证方式获取SAMLResponse，并采用浏览器提交表单数据的方式，获取unscoped token。
func (c *IamClient) CreateUnscopeTokenByIdpInitiated(request *model.CreateUnscopeTokenByIdpInitiatedRequest) (*model.CreateUnscopeTokenByIdpInitiatedResponse, error) {
	requestDef := GenReqDefForCreateUnscopeTokenByIdpInitiated()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateUnscopeTokenByIdpInitiatedResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)删除委托。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) DeleteAgency(request *model.DeleteAgencyRequest) (*model.DeleteAgencyResponse, error) {
	requestDef := GenReqDefForDeleteAgency()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteAgencyResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)删除自定义策略。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) DeleteCustomPolicy(request *model.DeleteCustomPolicyRequest) (*model.DeleteCustomPolicyResponse, error) {
	requestDef := GenReqDefForDeleteCustomPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteCustomPolicyResponse), nil
	}
}

//该接口可以用于移除用户组的所有项目服务权限。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) DeleteDomainGroupInheritedRole(request *model.DeleteDomainGroupInheritedRoleRequest) (*model.DeleteDomainGroupInheritedRoleResponse, error) {
	requestDef := GenReqDefForDeleteDomainGroupInheritedRole()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteDomainGroupInheritedRoleResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)添加IAM用户到用户组。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneAddUserToGroup(request *model.KeystoneAddUserToGroupRequest) (*model.KeystoneAddUserToGroupResponse, error) {
	requestDef := GenReqDefForKeystoneAddUserToGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneAddUserToGroupResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)为用户组授予全局服务权限。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneAssociateGroupWithDomainPermission(request *model.KeystoneAssociateGroupWithDomainPermissionRequest) (*model.KeystoneAssociateGroupWithDomainPermissionResponse, error) {
	requestDef := GenReqDefForKeystoneAssociateGroupWithDomainPermission()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneAssociateGroupWithDomainPermissionResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)为用户组授予项目服务权限。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneAssociateGroupWithProjectPermission(request *model.KeystoneAssociateGroupWithProjectPermissionRequest) (*model.KeystoneAssociateGroupWithProjectPermissionResponse, error) {
	requestDef := GenReqDefForKeystoneAssociateGroupWithProjectPermission()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneAssociateGroupWithProjectPermissionResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询用户组是否拥有全局服务权限。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneCheckDomainPermissionForGroup(request *model.KeystoneCheckDomainPermissionForGroupRequest) (*model.KeystoneCheckDomainPermissionForGroupResponse, error) {
	requestDef := GenReqDefForKeystoneCheckDomainPermissionForGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneCheckDomainPermissionForGroupResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询用户组是否拥有项目服务权限。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneCheckProjectPermissionForGroup(request *model.KeystoneCheckProjectPermissionForGroupRequest) (*model.KeystoneCheckProjectPermissionForGroupResponse, error) {
	requestDef := GenReqDefForKeystoneCheckProjectPermissionForGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneCheckProjectPermissionForGroupResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询IAM用户是否在用户组中。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneCheckUserInGroup(request *model.KeystoneCheckUserInGroupRequest) (*model.KeystoneCheckUserInGroupResponse, error) {
	requestDef := GenReqDefForKeystoneCheckUserInGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneCheckUserInGroupResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询用户组是否拥有所有项目指定权限。  该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneCheckroleForGroup(request *model.KeystoneCheckroleForGroupRequest) (*model.KeystoneCheckroleForGroupResponse, error) {
	requestDef := GenReqDefForKeystoneCheckroleForGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneCheckroleForGroupResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)创建用户组。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneCreateGroup(request *model.KeystoneCreateGroupRequest) (*model.KeystoneCreateGroupResponse, error) {
	requestDef := GenReqDefForKeystoneCreateGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneCreateGroupResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)注册身份提供商。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneCreateIdentityProvider(request *model.KeystoneCreateIdentityProviderRequest) (*model.KeystoneCreateIdentityProviderResponse, error) {
	requestDef := GenReqDefForKeystoneCreateIdentityProvider()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneCreateIdentityProviderResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)注册映射。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneCreateMapping(request *model.KeystoneCreateMappingRequest) (*model.KeystoneCreateMappingResponse, error) {
	requestDef := GenReqDefForKeystoneCreateMapping()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneCreateMappingResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)创建项目。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneCreateProject(request *model.KeystoneCreateProjectRequest) (*model.KeystoneCreateProjectResponse, error) {
	requestDef := GenReqDefForKeystoneCreateProject()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneCreateProjectResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)注册协议（将协议关联到某一身份提供商）。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneCreateProtocol(request *model.KeystoneCreateProtocolRequest) (*model.KeystoneCreateProtocolResponse, error) {
	requestDef := GenReqDefForKeystoneCreateProtocol()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneCreateProtocolResponse), nil
	}
}

//该接口可以用于通过联邦认证方式获取scoped token。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneCreateScopedToken(request *model.KeystoneCreateScopedTokenRequest) (*model.KeystoneCreateScopedTokenResponse, error) {
	requestDef := GenReqDefForKeystoneCreateScopedToken()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneCreateScopedTokenResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)删除用户组。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneDeleteGroup(request *model.KeystoneDeleteGroupRequest) (*model.KeystoneDeleteGroupResponse, error) {
	requestDef := GenReqDefForKeystoneDeleteGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneDeleteGroupResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html) 删除身份提供商。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneDeleteIdentityProvider(request *model.KeystoneDeleteIdentityProviderRequest) (*model.KeystoneDeleteIdentityProviderResponse, error) {
	requestDef := GenReqDefForKeystoneDeleteIdentityProvider()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneDeleteIdentityProviderResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)删除映射。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneDeleteMapping(request *model.KeystoneDeleteMappingRequest) (*model.KeystoneDeleteMappingResponse, error) {
	requestDef := GenReqDefForKeystoneDeleteMapping()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneDeleteMappingResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)删除协议。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneDeleteProtocol(request *model.KeystoneDeleteProtocolRequest) (*model.KeystoneDeleteProtocolResponse, error) {
	requestDef := GenReqDefForKeystoneDeleteProtocol()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneDeleteProtocolResponse), nil
	}
}

//该接口可以用于管理员查询用户组所有项目服务权限列表。  \\n\\n该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneListAllProjectPermissionsForGroup(request *model.KeystoneListAllProjectPermissionsForGroupRequest) (*model.KeystoneListAllProjectPermissionsForGroupResponse, error) {
	requestDef := GenReqDefForKeystoneListAllProjectPermissionsForGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneListAllProjectPermissionsForGroupResponse), nil
	}
}

//该接口可以用于查询IAM用户可以用访问的账号详情。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneListAuthDomains(request *model.KeystoneListAuthDomainsRequest) (*model.KeystoneListAuthDomainsResponse, error) {
	requestDef := GenReqDefForKeystoneListAuthDomains()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneListAuthDomainsResponse), nil
	}
}

//该接口可以用于查询IAM用户可以访问的项目列表。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneListAuthProjects(request *model.KeystoneListAuthProjectsRequest) (*model.KeystoneListAuthProjectsResponse, error) {
	requestDef := GenReqDefForKeystoneListAuthProjects()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneListAuthProjectsResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询全局服务中的用户组权限。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneListDomainPermissionsForGroup(request *model.KeystoneListDomainPermissionsForGroupRequest) (*model.KeystoneListDomainPermissionsForGroupResponse, error) {
	requestDef := GenReqDefForKeystoneListDomainPermissionsForGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneListDomainPermissionsForGroupResponse), nil
	}
}

//该接口可以用于查询终端节点列表。终端节点用来提供服务访问入口。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneListEndpoints(request *model.KeystoneListEndpointsRequest) (*model.KeystoneListEndpointsResponse, error) {
	requestDef := GenReqDefForKeystoneListEndpoints()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneListEndpointsResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询用户组列表。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneListGroups(request *model.KeystoneListGroupsRequest) (*model.KeystoneListGroupsResponse, error) {
	requestDef := GenReqDefForKeystoneListGroups()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneListGroupsResponse), nil
	}
}

//该接口可以用于查询身份提供商列表。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneListIdentityProviders(request *model.KeystoneListIdentityProvidersRequest) (*model.KeystoneListIdentityProvidersResponse, error) {
	requestDef := GenReqDefForKeystoneListIdentityProviders()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneListIdentityProvidersResponse), nil
	}
}

//该接口可以用于查询映射列表。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneListMappings(request *model.KeystoneListMappingsRequest) (*model.KeystoneListMappingsResponse, error) {
	requestDef := GenReqDefForKeystoneListMappings()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneListMappingsResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询权限列表。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneListPermissions(request *model.KeystoneListPermissionsRequest) (*model.KeystoneListPermissionsResponse, error) {
	requestDef := GenReqDefForKeystoneListPermissions()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneListPermissionsResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询项目服务中的用户组权限。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneListProjectPermissionsForGroup(request *model.KeystoneListProjectPermissionsForGroupRequest) (*model.KeystoneListProjectPermissionsForGroupResponse, error) {
	requestDef := GenReqDefForKeystoneListProjectPermissionsForGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneListProjectPermissionsForGroupResponse), nil
	}
}

//该接口可以用于查询指定条件下的项目列表。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneListProjects(request *model.KeystoneListProjectsRequest) (*model.KeystoneListProjectsResponse, error) {
	requestDef := GenReqDefForKeystoneListProjects()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneListProjectsResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询指定IAM用户的项目列表，或IAM用户查询自己的项目列表。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneListProjectsForUser(request *model.KeystoneListProjectsForUserRequest) (*model.KeystoneListProjectsForUserResponse, error) {
	requestDef := GenReqDefForKeystoneListProjectsForUser()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneListProjectsForUserResponse), nil
	}
}

//该接口可以用于查询协议列表。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneListProtocols(request *model.KeystoneListProtocolsRequest) (*model.KeystoneListProtocolsResponse, error) {
	requestDef := GenReqDefForKeystoneListProtocols()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneListProtocolsResponse), nil
	}
}

//该接口可以用于查询区域列表。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneListRegions(request *model.KeystoneListRegionsRequest) (*model.KeystoneListRegionsResponse, error) {
	requestDef := GenReqDefForKeystoneListRegions()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneListRegionsResponse), nil
	}
}

//该接口可以用于查询服务列表。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneListServices(request *model.KeystoneListServicesRequest) (*model.KeystoneListServicesResponse, error) {
	requestDef := GenReqDefForKeystoneListServices()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneListServicesResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询用户组中所包含的IAM用户。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneListUsersForGroupByAdmin(request *model.KeystoneListUsersForGroupByAdminRequest) (*model.KeystoneListUsersForGroupByAdminResponse, error) {
	requestDef := GenReqDefForKeystoneListUsersForGroupByAdmin()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneListUsersForGroupByAdminResponse), nil
	}
}

//该接口用于查询Keystone API的版本信息。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneListVersions(request *model.KeystoneListVersionsRequest) (*model.KeystoneListVersionsResponse, error) {
	requestDef := GenReqDefForKeystoneListVersions()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneListVersionsResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)移除用户组的全局服务权限。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneRemoveDomainPermissionFromGroup(request *model.KeystoneRemoveDomainPermissionFromGroupRequest) (*model.KeystoneRemoveDomainPermissionFromGroupResponse, error) {
	requestDef := GenReqDefForKeystoneRemoveDomainPermissionFromGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneRemoveDomainPermissionFromGroupResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)移除用户组的项目服务权限。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneRemoveProjectPermissionFromGroup(request *model.KeystoneRemoveProjectPermissionFromGroupRequest) (*model.KeystoneRemoveProjectPermissionFromGroupResponse, error) {
	requestDef := GenReqDefForKeystoneRemoveProjectPermissionFromGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneRemoveProjectPermissionFromGroupResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)移除用户组中的IAM用户。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneRemoveUserFromGroup(request *model.KeystoneRemoveUserFromGroupRequest) (*model.KeystoneRemoveUserFromGroupResponse, error) {
	requestDef := GenReqDefForKeystoneRemoveUserFromGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneRemoveUserFromGroupResponse), nil
	}
}

//该接口可以用于查询请求头中X-Auth-Token对应的服务目录。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneShowCatalog(request *model.KeystoneShowCatalogRequest) (*model.KeystoneShowCatalogResponse, error) {
	requestDef := GenReqDefForKeystoneShowCatalog()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneShowCatalogResponse), nil
	}
}

//该接口可以用于查询终端节点详情。终端节点用来提供服务访问入口。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneShowEndpoint(request *model.KeystoneShowEndpointRequest) (*model.KeystoneShowEndpointResponse, error) {
	requestDef := GenReqDefForKeystoneShowEndpoint()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneShowEndpointResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询用户组详情。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneShowGroup(request *model.KeystoneShowGroupRequest) (*model.KeystoneShowGroupResponse, error) {
	requestDef := GenReqDefForKeystoneShowGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneShowGroupResponse), nil
	}
}

//该接口可以用于查询身份提供商详情。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneShowIdentityProvider(request *model.KeystoneShowIdentityProviderRequest) (*model.KeystoneShowIdentityProviderResponse, error) {
	requestDef := GenReqDefForKeystoneShowIdentityProvider()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneShowIdentityProviderResponse), nil
	}
}

//该接口可以用于查询映射详情。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneShowMapping(request *model.KeystoneShowMappingRequest) (*model.KeystoneShowMappingResponse, error) {
	requestDef := GenReqDefForKeystoneShowMapping()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneShowMappingResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询权限详情。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneShowPermission(request *model.KeystoneShowPermissionRequest) (*model.KeystoneShowPermissionResponse, error) {
	requestDef := GenReqDefForKeystoneShowPermission()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneShowPermissionResponse), nil
	}
}

//该接口可以用于查询项目详情。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneShowProject(request *model.KeystoneShowProjectRequest) (*model.KeystoneShowProjectResponse, error) {
	requestDef := GenReqDefForKeystoneShowProject()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneShowProjectResponse), nil
	}
}

//该接口可以用于查询协议详情。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneShowProtocol(request *model.KeystoneShowProtocolRequest) (*model.KeystoneShowProtocolResponse, error) {
	requestDef := GenReqDefForKeystoneShowProtocol()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneShowProtocolResponse), nil
	}
}

//该接口可以用于查询区域详情。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneShowRegion(request *model.KeystoneShowRegionRequest) (*model.KeystoneShowRegionResponse, error) {
	requestDef := GenReqDefForKeystoneShowRegion()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneShowRegionResponse), nil
	}
}

//该接口可以用于查询账号密码强度策略，查询结果包括密码强度策略的正则表达式及其描述。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneShowSecurityCompliance(request *model.KeystoneShowSecurityComplianceRequest) (*model.KeystoneShowSecurityComplianceResponse, error) {
	requestDef := GenReqDefForKeystoneShowSecurityCompliance()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneShowSecurityComplianceResponse), nil
	}
}

//该接口可以用于按条件查询账号密码强度策略，查询结果包括密码强度策略的正则表达式及其描述。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneShowSecurityComplianceByOption(request *model.KeystoneShowSecurityComplianceByOptionRequest) (*model.KeystoneShowSecurityComplianceByOptionResponse, error) {
	requestDef := GenReqDefForKeystoneShowSecurityComplianceByOption()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneShowSecurityComplianceByOptionResponse), nil
	}
}

//该接口可以用于查询服务详情。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneShowService(request *model.KeystoneShowServiceRequest) (*model.KeystoneShowServiceResponse, error) {
	requestDef := GenReqDefForKeystoneShowService()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneShowServiceResponse), nil
	}
}

//该接口用于查询Keystone API的3.0版本的信息。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneShowVersion(request *model.KeystoneShowVersionRequest) (*model.KeystoneShowVersionResponse, error) {
	requestDef := GenReqDefForKeystoneShowVersion()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneShowVersionResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)更新用户组信息。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneUpdateGroup(request *model.KeystoneUpdateGroupRequest) (*model.KeystoneUpdateGroupResponse, error) {
	requestDef := GenReqDefForKeystoneUpdateGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneUpdateGroupResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)更新身份提供商。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneUpdateIdentityProvider(request *model.KeystoneUpdateIdentityProviderRequest) (*model.KeystoneUpdateIdentityProviderResponse, error) {
	requestDef := GenReqDefForKeystoneUpdateIdentityProvider()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneUpdateIdentityProviderResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)更新映射。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneUpdateMapping(request *model.KeystoneUpdateMappingRequest) (*model.KeystoneUpdateMappingResponse, error) {
	requestDef := GenReqDefForKeystoneUpdateMapping()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneUpdateMappingResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)修改项目信息。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneUpdateProject(request *model.KeystoneUpdateProjectRequest) (*model.KeystoneUpdateProjectResponse, error) {
	requestDef := GenReqDefForKeystoneUpdateProject()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneUpdateProjectResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)更新协议。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneUpdateProtocol(request *model.KeystoneUpdateProtocolRequest) (*model.KeystoneUpdateProtocolResponse, error) {
	requestDef := GenReqDefForKeystoneUpdateProtocol()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneUpdateProtocolResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询指定条件下的委托列表。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) ListAgencies(request *model.ListAgenciesRequest) (*model.ListAgenciesResponse, error) {
	requestDef := GenReqDefForListAgencies()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListAgenciesResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询委托所有项目服务权限列表。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) ListAllProjectsPermissionsForAgency(request *model.ListAllProjectsPermissionsForAgencyRequest) (*model.ListAllProjectsPermissionsForAgencyResponse, error) {
	requestDef := GenReqDefForListAllProjectsPermissionsForAgency()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListAllProjectsPermissionsForAgencyResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询自定义策略列表。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) ListCustomPolicies(request *model.ListCustomPoliciesRequest) (*model.ListCustomPoliciesResponse, error) {
	requestDef := GenReqDefForListCustomPolicies()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListCustomPoliciesResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询全局服务中的委托权限。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) ListDomainPermissionsForAgency(request *model.ListDomainPermissionsForAgencyRequest) (*model.ListDomainPermissionsForAgencyResponse, error) {
	requestDef := GenReqDefForListDomainPermissionsForAgency()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListDomainPermissionsForAgencyResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询项目服务中的委托权限。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) ListProjectPermissionsForAgency(request *model.ListProjectPermissionsForAgencyRequest) (*model.ListProjectPermissionsForAgencyResponse, error) {
	requestDef := GenReqDefForListProjectPermissionsForAgency()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListProjectPermissionsForAgencyResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)移除委托的所有项目服务权限。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) RemoveAllProjectsPermissionFromAgency(request *model.RemoveAllProjectsPermissionFromAgencyRequest) (*model.RemoveAllProjectsPermissionFromAgencyResponse, error) {
	requestDef := GenReqDefForRemoveAllProjectsPermissionFromAgency()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.RemoveAllProjectsPermissionFromAgencyResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)移除委托的全局服务权限。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) RemoveDomainPermissionFromAgency(request *model.RemoveDomainPermissionFromAgencyRequest) (*model.RemoveDomainPermissionFromAgencyResponse, error) {
	requestDef := GenReqDefForRemoveDomainPermissionFromAgency()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.RemoveDomainPermissionFromAgencyResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)移除委托的项目服务权限。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) RemoveProjectPermissionFromAgency(request *model.RemoveProjectPermissionFromAgencyRequest) (*model.RemoveProjectPermissionFromAgencyResponse, error) {
	requestDef := GenReqDefForRemoveProjectPermissionFromAgency()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.RemoveProjectPermissionFromAgencyResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询委托详情。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) ShowAgency(request *model.ShowAgencyRequest) (*model.ShowAgencyResponse, error) {
	requestDef := GenReqDefForShowAgency()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowAgencyResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询自定义策略详情。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) ShowCustomPolicy(request *model.ShowCustomPolicyRequest) (*model.ShowCustomPolicyResponse, error) {
	requestDef := GenReqDefForShowCustomPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowCustomPolicyResponse), nil
	}
}

//该接口可以用于查询账号接口访问控制策略。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) ShowDomainApiAclPolicy(request *model.ShowDomainApiAclPolicyRequest) (*model.ShowDomainApiAclPolicyResponse, error) {
	requestDef := GenReqDefForShowDomainApiAclPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowDomainApiAclPolicyResponse), nil
	}
}

//该接口可以用于查询账号控制台访问控制策略。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) ShowDomainConsoleAclPolicy(request *model.ShowDomainConsoleAclPolicyRequest) (*model.ShowDomainConsoleAclPolicyResponse, error) {
	requestDef := GenReqDefForShowDomainConsoleAclPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowDomainConsoleAclPolicyResponse), nil
	}
}

//该接口可以用于查询账号登录策略。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) ShowDomainLoginPolicy(request *model.ShowDomainLoginPolicyRequest) (*model.ShowDomainLoginPolicyResponse, error) {
	requestDef := GenReqDefForShowDomainLoginPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowDomainLoginPolicyResponse), nil
	}
}

//该接口可以用于查询账号密码策略。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) ShowDomainPasswordPolicy(request *model.ShowDomainPasswordPolicyRequest) (*model.ShowDomainPasswordPolicyResponse, error) {
	requestDef := GenReqDefForShowDomainPasswordPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowDomainPasswordPolicyResponse), nil
	}
}

//该接口可以用于查询账号操作保护策略。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) ShowDomainProtectPolicy(request *model.ShowDomainProtectPolicyRequest) (*model.ShowDomainProtectPolicyResponse, error) {
	requestDef := GenReqDefForShowDomainProtectPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowDomainProtectPolicyResponse), nil
	}
}

//该接口可以用于查询账号配额。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) ShowDomainQuota(request *model.ShowDomainQuotaRequest) (*model.ShowDomainQuotaResponse, error) {
	requestDef := GenReqDefForShowDomainQuota()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowDomainQuotaResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询身份提供商导入到IAM中的Metadata文件。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) ShowMetadata(request *model.ShowMetadataRequest) (*model.ShowMetadataResponse, error) {
	requestDef := GenReqDefForShowMetadata()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowMetadataResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询项目详情与状态。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) ShowProjectDetailsAndStatus(request *model.ShowProjectDetailsAndStatusRequest) (*model.ShowProjectDetailsAndStatusResponse, error) {
	requestDef := GenReqDefForShowProjectDetailsAndStatus()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowProjectDetailsAndStatusResponse), nil
	}
}

//该接口可以用于查询项目配额。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) ShowProjectQuota(request *model.ShowProjectQuotaRequest) (*model.ShowProjectQuotaResponse, error) {
	requestDef := GenReqDefForShowProjectQuota()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowProjectQuotaResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)修改委托。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) UpdateAgency(request *model.UpdateAgencyRequest) (*model.UpdateAgencyResponse, error) {
	requestDef := GenReqDefForUpdateAgency()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateAgencyResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)修改委托自定义策略。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) UpdateAgencyCustomPolicy(request *model.UpdateAgencyCustomPolicyRequest) (*model.UpdateAgencyCustomPolicyResponse, error) {
	requestDef := GenReqDefForUpdateAgencyCustomPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateAgencyCustomPolicyResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)修改云服务自定义策略。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) UpdateCloudServiceCustomPolicy(request *model.UpdateCloudServiceCustomPolicyRequest) (*model.UpdateCloudServiceCustomPolicyResponse, error) {
	requestDef := GenReqDefForUpdateCloudServiceCustomPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateCloudServiceCustomPolicyResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)修改账号接口访问策略。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) UpdateDomainApiAclPolicy(request *model.UpdateDomainApiAclPolicyRequest) (*model.UpdateDomainApiAclPolicyResponse, error) {
	requestDef := GenReqDefForUpdateDomainApiAclPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateDomainApiAclPolicyResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)修改账号控制台访问策略。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) UpdateDomainConsoleAclPolicy(request *model.UpdateDomainConsoleAclPolicyRequest) (*model.UpdateDomainConsoleAclPolicyResponse, error) {
	requestDef := GenReqDefForUpdateDomainConsoleAclPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateDomainConsoleAclPolicyResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/zh-cn_topic_0079496985.html)为用户组授予所有项目服务权限。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) UpdateDomainGroupInheritRole(request *model.UpdateDomainGroupInheritRoleRequest) (*model.UpdateDomainGroupInheritRoleResponse, error) {
	requestDef := GenReqDefForUpdateDomainGroupInheritRole()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateDomainGroupInheritRoleResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)修改账号登录策略。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) UpdateDomainLoginPolicy(request *model.UpdateDomainLoginPolicyRequest) (*model.UpdateDomainLoginPolicyResponse, error) {
	requestDef := GenReqDefForUpdateDomainLoginPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateDomainLoginPolicyResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)修改账号密码策略。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) UpdateDomainPasswordPolicy(request *model.UpdateDomainPasswordPolicyRequest) (*model.UpdateDomainPasswordPolicyResponse, error) {
	requestDef := GenReqDefForUpdateDomainPasswordPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateDomainPasswordPolicyResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)修改账号操作保护策略。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) UpdateDomainProtectPolicy(request *model.UpdateDomainProtectPolicyRequest) (*model.UpdateDomainProtectPolicyResponse, error) {
	requestDef := GenReqDefForUpdateDomainProtectPolicy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateDomainProtectPolicyResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)设置项目状态。项目状态包括：正常、冻结。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) UpdateProjectStatus(request *model.UpdateProjectStatusRequest) (*model.UpdateProjectStatusResponse, error) {
	requestDef := GenReqDefForUpdateProjectStatus()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateProjectStatusResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)给IAM用户创建永久访问密钥，或IAM用户给自己创建永久访问密钥。    访问密钥（Access Key ID/Secret Access Key，简称AK/SK），是您通过开发工具（API、CLI、SDK）访问华为云时的身份凭证，不用于登录控制台。系统通过AK识别访问用户的身份，通过SK进行签名验证，通过加密签名验证可以确保请求的机密性、完整性和请求者身份的正确性。在控制台创建访问密钥的方式请参见：[访问密钥](https://support.huaweicloud.com/usermanual-ca/zh-cn_topic_0046606340.html)  。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) CreatePermanentAccessKey(request *model.CreatePermanentAccessKeyRequest) (*model.CreatePermanentAccessKeyResponse, error) {
	requestDef := GenReqDefForCreatePermanentAccessKey()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreatePermanentAccessKeyResponse), nil
	}
}

//该接口可以用于通过委托来获取临时访问密钥（临时AK/SK）和securitytoken。    临时AK/SK和securitytoken是系统颁发给IAM用户的临时访问令牌，有效期为15分钟至24小时，过期后需要重新获取。临时AK/SK和securitytoken遵循权限最小化原则。鉴权时，临时AK/SK和securitytoken必须同时使用，请求头中需要添加“x-security-token”字段，使用方法详情请参考：[API签名参考](https://support.huaweicloud.com/devg-apisign/api-sign-provide.html) 。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) CreateTemporaryAccessKeyByAgency(request *model.CreateTemporaryAccessKeyByAgencyRequest) (*model.CreateTemporaryAccessKeyByAgencyResponse, error) {
	requestDef := GenReqDefForCreateTemporaryAccessKeyByAgency()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateTemporaryAccessKeyByAgencyResponse), nil
	}
}

//该接口可以用于通过token来获取临时AK/SK和securitytoken。    临时AK/SK和securitytoken是系统颁发给IAM用户的临时访问令牌，有效期为15分钟至24小时，过期后需要重新获取。临时AK/SK和securitytoken遵循权限最小化原则。鉴权时，临时AK/SK和securitytoken必须同时使用，请求头中需要添加“x-security-token”字段，使用方法详情请参考：[API签名参考](https://support.huaweicloud.com/devg-apisign/api-sign-provide.html)。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) CreateTemporaryAccessKeyByToken(request *model.CreateTemporaryAccessKeyByTokenRequest) (*model.CreateTemporaryAccessKeyByTokenResponse, error) {
	requestDef := GenReqDefForCreateTemporaryAccessKeyByToken()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateTemporaryAccessKeyByTokenResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)删除IAM用户的指定永久访问密钥，或IAM用户删除自己的指定永久访问密钥。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) DeletePermanentAccessKey(request *model.DeletePermanentAccessKeyRequest) (*model.DeletePermanentAccessKeyResponse, error) {
	requestDef := GenReqDefForDeletePermanentAccessKey()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeletePermanentAccessKeyResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询IAM用户的所有永久访问密钥，或IAM用户查询自己的所有永久访问密钥。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) ListPermanentAccessKeys(request *model.ListPermanentAccessKeysRequest) (*model.ListPermanentAccessKeysResponse, error) {
	requestDef := GenReqDefForListPermanentAccessKeys()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListPermanentAccessKeysResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询IAM用户的指定永久访问密钥，或IAM用户查询自己的指定永久访问密钥。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) ShowPermanentAccessKey(request *model.ShowPermanentAccessKeyRequest) (*model.ShowPermanentAccessKeyResponse, error) {
	requestDef := GenReqDefForShowPermanentAccessKey()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowPermanentAccessKeyResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)修改IAM用户的指定永久访问密钥，或IAM用户修改自己的指定永久访问密钥。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) UpdatePermanentAccessKey(request *model.UpdatePermanentAccessKeyRequest) (*model.UpdatePermanentAccessKeyResponse, error) {
	requestDef := GenReqDefForUpdatePermanentAccessKey()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdatePermanentAccessKeyResponse), nil
	}
}

//该接口可以用于绑定MFA设备。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) CreateBindingDevice(request *model.CreateBindingDeviceRequest) (*model.CreateBindingDeviceResponse, error) {
	requestDef := GenReqDefForCreateBindingDevice()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateBindingDeviceResponse), nil
	}
}

//该接口可以用于创建MFA设备。  该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) CreateMfaDevice(request *model.CreateMfaDeviceRequest) (*model.CreateMfaDeviceResponse, error) {
	requestDef := GenReqDefForCreateMfaDevice()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateMfaDeviceResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)创建IAM用户。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) CreateUser(request *model.CreateUserRequest) (*model.CreateUserResponse, error) {
	requestDef := GenReqDefForCreateUser()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateUserResponse), nil
	}
}

//该接口可以用于解绑MFA设备   该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) DeleteBindingDevice(request *model.DeleteBindingDeviceRequest) (*model.DeleteBindingDeviceResponse, error) {
	requestDef := GenReqDefForDeleteBindingDevice()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteBindingDeviceResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)删除MFA设备。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) DeleteMfaDevice(request *model.DeleteMfaDeviceRequest) (*model.DeleteMfaDeviceResponse, error) {
	requestDef := GenReqDefForDeleteMfaDevice()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteMfaDeviceResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)创建IAM用户。IAM用户首次登录时需要修改密码。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneCreateUser(request *model.KeystoneCreateUserRequest) (*model.KeystoneCreateUserResponse, error) {
	requestDef := GenReqDefForKeystoneCreateUser()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneCreateUserResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)删除指定IAM用户。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneDeleteUser(request *model.KeystoneDeleteUserRequest) (*model.KeystoneDeleteUserResponse, error) {
	requestDef := GenReqDefForKeystoneDeleteUser()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneDeleteUserResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询IAM用户所属用户组，或IAM用户查询自己所属用户组。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneListGroupsForUser(request *model.KeystoneListGroupsForUserRequest) (*model.KeystoneListGroupsForUserResponse, error) {
	requestDef := GenReqDefForKeystoneListGroupsForUser()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneListGroupsForUserResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询IAM用户列表。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneListUsers(request *model.KeystoneListUsersRequest) (*model.KeystoneListUsersResponse, error) {
	requestDef := GenReqDefForKeystoneListUsers()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneListUsersResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询IAM用户详情，或IAM用户查询自己的用户详情。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneShowUser(request *model.KeystoneShowUserRequest) (*model.KeystoneShowUserResponse, error) {
	requestDef := GenReqDefForKeystoneShowUser()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneShowUserResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)修改IAM用户信息。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneUpdateUserByAdmin(request *model.KeystoneUpdateUserByAdminRequest) (*model.KeystoneUpdateUserByAdminResponse, error) {
	requestDef := GenReqDefForKeystoneUpdateUserByAdmin()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneUpdateUserByAdminResponse), nil
	}
}

//该接口可以用于IAM用户修改自己的密码。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneUpdateUserPassword(request *model.KeystoneUpdateUserPasswordRequest) (*model.KeystoneUpdateUserPasswordResponse, error) {
	requestDef := GenReqDefForKeystoneUpdateUserPassword()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneUpdateUserPasswordResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询IAM用户的登录保护状态列表。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) ListUserLoginProtects(request *model.ListUserLoginProtectsRequest) (*model.ListUserLoginProtectsResponse, error) {
	requestDef := GenReqDefForListUserLoginProtects()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListUserLoginProtectsResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询IAM用户的MFA绑定信息列表。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) ListUserMfaDevices(request *model.ListUserMfaDevicesRequest) (*model.ListUserMfaDevicesResponse, error) {
	requestDef := GenReqDefForListUserMfaDevices()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListUserMfaDevicesResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询IAM用户详情，或IAM用户查询自己的详情。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) ShowUser(request *model.ShowUserRequest) (*model.ShowUserResponse, error) {
	requestDef := GenReqDefForShowUser()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowUserResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询指定IAM用户的登录保护状态信息，或IAM用户查询自己的登录保护状态信息。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) ShowUserLoginProtect(request *model.ShowUserLoginProtectRequest) (*model.ShowUserLoginProtectResponse, error) {
	requestDef := GenReqDefForShowUserLoginProtect()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowUserLoginProtectResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)查询指定IAM用户的MFA绑定信息，或IAM用户查询自己的MFA绑定信息。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) ShowUserMfaDevice(request *model.ShowUserMfaDeviceRequest) (*model.ShowUserMfaDeviceResponse, error) {
	requestDef := GenReqDefForShowUserMfaDevice()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowUserMfaDeviceResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)修改账号操作保护。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) UpdateLoginProtect(request *model.UpdateLoginProtectRequest) (*model.UpdateLoginProtectResponse, error) {
	requestDef := GenReqDefForUpdateLoginProtect()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateLoginProtectResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)修改IAM用户信息 。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) UpdateUser(request *model.UpdateUserRequest) (*model.UpdateUserResponse, error) {
	requestDef := GenReqDefForUpdateUser()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateUserResponse), nil
	}
}

//该接口可以用于IAM用户修改自己的用户信息。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) UpdateUserInformation(request *model.UpdateUserInformationRequest) (*model.UpdateUserInformationResponse, error) {
	requestDef := GenReqDefForUpdateUserInformation()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateUserInformationResponse), nil
	}
}

//该接口可以用于获取委托方的token。    例如：A账号希望B账号管理自己的某些资源，所以A账号创建了委托给B账号，则A账号为委托方，B账号为被委托方。那么B账号可以通过该接口获取委托token。B账号仅能使用该token管理A账号的委托资源，不能管理自己账号中的资源。如果B账号需要管理自己账号中的资源，则需要获取自己的用户token。    token是系统颁发给用户的访问令牌，承载用户的身份、权限等信息。调用IAM以及其他云服务的接口时，可以使用本接口获取的token进行鉴权。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。如果使用全局区域的Endpoint调用，该token可以在所有区域使用；如果使用非全局区域的Endpoint调用，则该token仅在该区域生效，不能跨区域使用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。    > - token的有效期为24小时，建议进行缓存，避免频繁调用。
func (c *IamClient) KeystoneCreateAgencyToken(request *model.KeystoneCreateAgencyTokenRequest) (*model.KeystoneCreateAgencyTokenResponse, error) {
	requestDef := GenReqDefForKeystoneCreateAgencyToken()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneCreateAgencyTokenResponse), nil
	}
}

//该接口可以用于通过用户名/密码的方式进行认证来获取IAM用户token。    token是系统颁发给IAM用户的访问令牌，承载用户的身份、权限等信息。调用IAM以及其他云服务的接口时，可以使用本接口获取的IAM用户token进行鉴权。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。如果使用全局区域的Endpoint调用，该token可以在所有区域使用；如果使用非全局区域的Endpoint调用，则该token仅在该区域生效，不能跨区域使用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。   > - token的有效期为24小时，建议进行缓存，避免频繁调用。   > - 通过Postman获取用户token示例请参见：[如何通过Postman获取用户token](https://support.huaweicloud.com/iam_faq/iam_01_034.html)。   > - 如果需要获取具有Security Administrator权限的token，请参见：[IAM 常见问题](https://support.huaweicloud.com/iam_faq/iam_01_0608.html)。
func (c *IamClient) KeystoneCreateUserTokenByPassword(request *model.KeystoneCreateUserTokenByPasswordRequest) (*model.KeystoneCreateUserTokenByPasswordResponse, error) {
	requestDef := GenReqDefForKeystoneCreateUserTokenByPassword()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneCreateUserTokenByPasswordResponse), nil
	}
}

//该接口可以用于通过用户名/密码+虚拟MFA的方式进行认证，在IAM用户开启了的登录保护功能，并选择通过虚拟MFA验证时获取IAM用户token。    token是系统颁发给用户的访问令牌，承载用户的身份、权限等信息。调用IAM以及其他云服务的接口时，可以使用本接口获取的token进行鉴权。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。如果使用全局区域的Endpoint调用，该token可以在所有区域使用；如果使用非全局区域的Endpoint调用，则该token仅在该区域生效，不能跨区域使用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。   > - token的有效期为24小时，建议进行缓存，避免频繁调用。   > - 通过Postman获取用户token示例请参见：[如何通过Postman获取用户token](https://support.huaweicloud.com/iam_faq/iam_01_034.html)。   > - 如果需要获取具有Security Administrator权限的token，请参见：[IAM 常见问题](https://support.huaweicloud.com/iam_faq/iam_01_0608.html)。
func (c *IamClient) KeystoneCreateUserTokenByPasswordAndMfa(request *model.KeystoneCreateUserTokenByPasswordAndMfaRequest) (*model.KeystoneCreateUserTokenByPasswordAndMfaResponse, error) {
	requestDef := GenReqDefForKeystoneCreateUserTokenByPasswordAndMfa()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneCreateUserTokenByPasswordAndMfaResponse), nil
	}
}

//该接口可以用于[管理员](https://support.huaweicloud.com/usermanual-iam/iam_01_0001.html)校验本账号中IAM用户token的有效性，或IAM用户校验自己token的有效性。管理员仅能校验本账号中IAM用户token的有效性，不能校验其他账号中IAM用户token的有效性。如果被校验的token有效，则返回该token的详细信息。    该接口可以使用全局区域的Endpoint和其他区域的Endpoint调用。IAM的Endpoint请参见：[地区和终端节点](https://developer.huaweicloud.com/endpoint?IAM)。
func (c *IamClient) KeystoneValidateToken(request *model.KeystoneValidateTokenRequest) (*model.KeystoneValidateTokenResponse, error) {
	requestDef := GenReqDefForKeystoneValidateToken()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.KeystoneValidateTokenResponse), nil
	}
}
