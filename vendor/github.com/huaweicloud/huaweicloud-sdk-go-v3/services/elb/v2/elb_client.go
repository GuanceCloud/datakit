package v2

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v2/model"
)

type ElbClient struct {
	HcClient *http_client.HcHttpClient
}

func NewElbClient(hcClient *http_client.HcHttpClient) *ElbClient {
	return &ElbClient{HcClient: hcClient}
}

func ElbClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//给后端云服务器组添加健康检查
func (c *ElbClient) CreateHealthmonitor(request *model.CreateHealthmonitorRequest) (*model.CreateHealthmonitorResponse, error) {
	requestDef := GenReqDefForCreateHealthmonitor()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateHealthmonitorResponse), nil
	}
}

//创建listener关联的转发策略
func (c *ElbClient) CreateL7policy(request *model.CreateL7policyRequest) (*model.CreateL7policyResponse, error) {
	requestDef := GenReqDefForCreateL7policy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateL7policyResponse), nil
	}
}

//创建转发规则
func (c *ElbClient) CreateL7rule(request *model.CreateL7ruleRequest) (*model.CreateL7ruleResponse, error) {
	requestDef := GenReqDefForCreateL7rule()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateL7ruleResponse), nil
	}
}

//创建与负载均衡器绑定的监听器。
func (c *ElbClient) CreateListener(request *model.CreateListenerRequest) (*model.CreateListenerResponse, error) {
	requestDef := GenReqDefForCreateListener()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateListenerResponse), nil
	}
}

//创建私网类型的增强型负载均衡器。创建成功后，该接口会返回创建的增强型负载均衡器的ID、所属子网ID、负载均衡器IP等详细信息。若要创建公网类型的增强型负载均衡器，还需调用创建浮动IP的接口，将浮动IP与私网负载均衡器的vip_port_id绑定。
func (c *ElbClient) CreateLoadbalancer(request *model.CreateLoadbalancerRequest) (*model.CreateLoadbalancerResponse, error) {
	requestDef := GenReqDefForCreateLoadbalancer()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateLoadbalancerResponse), nil
	}
}

//添加属于某个后端云服务器组的后端云服务器。
func (c *ElbClient) CreateMember(request *model.CreateMemberRequest) (*model.CreateMemberResponse, error) {
	requestDef := GenReqDefForCreateMember()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateMemberResponse), nil
	}
}

//创建后端云服务器组。将多个后端云服务器添加到后端云服务器组中后，请求会在后端云服务器间按后端云服务器组的负载均衡算法和后端云服务器的权重来做请求分发。
func (c *ElbClient) CreatePool(request *model.CreatePoolRequest) (*model.CreatePoolResponse, error) {
	requestDef := GenReqDefForCreatePool()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreatePoolResponse), nil
	}
}

//创建白名单，控制监听器的访问权限。若开启了白名单功能，只有白名单中放通的IP可以访问该监听器的后端服务。
func (c *ElbClient) CreateWhitelist(request *model.CreateWhitelistRequest) (*model.CreateWhitelistResponse, error) {
	requestDef := GenReqDefForCreateWhitelist()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateWhitelistResponse), nil
	}
}

//删除健康检查
func (c *ElbClient) DeleteHealthmonitor(request *model.DeleteHealthmonitorRequest) (*model.DeleteHealthmonitorResponse, error) {
	requestDef := GenReqDefForDeleteHealthmonitor()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteHealthmonitorResponse), nil
	}
}

//删除转发策略
func (c *ElbClient) DeleteL7policy(request *model.DeleteL7policyRequest) (*model.DeleteL7policyResponse, error) {
	requestDef := GenReqDefForDeleteL7policy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteL7policyResponse), nil
	}
}

//删除转发规则
func (c *ElbClient) DeleteL7rule(request *model.DeleteL7ruleRequest) (*model.DeleteL7ruleResponse, error) {
	requestDef := GenReqDefForDeleteL7rule()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteL7ruleResponse), nil
	}
}

//根据指定ID删除监听器。提供级联删除选项，当选择级联删除时，删除和负载均衡器关联的转发规则、转发策略、白名单、标签等。
func (c *ElbClient) DeleteListener(request *model.DeleteListenerRequest) (*model.DeleteListenerResponse, error) {
	requestDef := GenReqDefForDeleteListener()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteListenerResponse), nil
	}
}

//根据指定ID删除负载均衡器。提供级联删除选项，当选择级联删除时，删除和负载均衡器关联的监听器、后端云服务器组、后端云服务器、健康检查、转发策略、转发规则、白名单、标签等
func (c *ElbClient) DeleteLoadbalancer(request *model.DeleteLoadbalancerRequest) (*model.DeleteLoadbalancerResponse, error) {
	requestDef := GenReqDefForDeleteLoadbalancer()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteLoadbalancerResponse), nil
	}
}

//删除后端云服务器
func (c *ElbClient) DeleteMember(request *model.DeleteMemberRequest) (*model.DeleteMemberResponse, error) {
	requestDef := GenReqDefForDeleteMember()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteMemberResponse), nil
	}
}

//删除后端云服务器组。
func (c *ElbClient) DeletePool(request *model.DeletePoolRequest) (*model.DeletePoolResponse, error) {
	requestDef := GenReqDefForDeletePool()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeletePoolResponse), nil
	}
}

//删除白名单
func (c *ElbClient) DeleteWhitelist(request *model.DeleteWhitelistRequest) (*model.DeleteWhitelistResponse, error) {
	requestDef := GenReqDefForDeleteWhitelist()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteWhitelistResponse), nil
	}
}

//查询健康检查列表
func (c *ElbClient) ListHealthmonitors(request *model.ListHealthmonitorsRequest) (*model.ListHealthmonitorsResponse, error) {
	requestDef := GenReqDefForListHealthmonitors()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListHealthmonitorsResponse), nil
	}
}

//查询转发策略。支持过滤查询和分页查询。
func (c *ElbClient) ListL7policies(request *model.ListL7policiesRequest) (*model.ListL7policiesResponse, error) {
	requestDef := GenReqDefForListL7policies()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListL7policiesResponse), nil
	}
}

//查询指定转发策略下关联的转发规则列表
func (c *ElbClient) ListL7rules(request *model.ListL7rulesRequest) (*model.ListL7rulesResponse, error) {
	requestDef := GenReqDefForListL7rules()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListL7rulesResponse), nil
	}
}

//查询监听器列表。支持过滤查询和分页查询。可以通过监听器ID、协议类型、监听端口号、关联的后端云服务器的IP等查询监听器。
func (c *ElbClient) ListListeners(request *model.ListListenersRequest) (*model.ListListenersResponse, error) {
	requestDef := GenReqDefForListListeners()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListListenersResponse), nil
	}
}

//查询负载均衡器列表。
func (c *ElbClient) ListLoadbalancers(request *model.ListLoadbalancersRequest) (*model.ListLoadbalancersResponse, error) {
	requestDef := GenReqDefForListLoadbalancers()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListLoadbalancersResponse), nil
	}
}

//添加属于某个后端云服务器组的后端云服务器。
func (c *ElbClient) ListMenbers(request *model.ListMenbersRequest) (*model.ListMenbersResponse, error) {
	requestDef := GenReqDefForListMenbers()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListMenbersResponse), nil
	}
}

//查询后端云服务器组列表。
func (c *ElbClient) ListPools(request *model.ListPoolsRequest) (*model.ListPoolsResponse, error) {
	requestDef := GenReqDefForListPools()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListPoolsResponse), nil
	}
}

//查询白名单，支持过滤查询和分页查询。
func (c *ElbClient) ListWhitelists(request *model.ListWhitelistsRequest) (*model.ListWhitelistsResponse, error) {
	requestDef := GenReqDefForListWhitelists()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListWhitelistsResponse), nil
	}
}

//根据指定ID查询健康检查详情。
func (c *ElbClient) ShowHealthmonitors(request *model.ShowHealthmonitorsRequest) (*model.ShowHealthmonitorsResponse, error) {
	requestDef := GenReqDefForShowHealthmonitors()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowHealthmonitorsResponse), nil
	}
}

//根据指定ID查询转发策略详情。
func (c *ElbClient) ShowL7policy(request *model.ShowL7policyRequest) (*model.ShowL7policyResponse, error) {
	requestDef := GenReqDefForShowL7policy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowL7policyResponse), nil
	}
}

//根据指定ID查询某转发策略下关联的转发规则详情。
func (c *ElbClient) ShowL7rule(request *model.ShowL7ruleRequest) (*model.ShowL7ruleResponse, error) {
	requestDef := GenReqDefForShowL7rule()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowL7ruleResponse), nil
	}
}

//根据指定ID查询监听器详情。
func (c *ElbClient) ShowListener(request *model.ShowListenerRequest) (*model.ShowListenerResponse, error) {
	requestDef := GenReqDefForShowListener()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowListenerResponse), nil
	}
}

//根据指定负载均衡器ID查询负载均衡器详情
func (c *ElbClient) ShowLoadbalancer(request *model.ShowLoadbalancerRequest) (*model.ShowLoadbalancerResponse, error) {
	requestDef := GenReqDefForShowLoadbalancer()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowLoadbalancerResponse), nil
	}
}

//查询负载均衡器状态树。可通过该接口查询负载均衡器关联的监听器、后端云服务器组、后端云服务器、健康检查、转发策略、转发规则的主要信息，了解负载均衡器下资源的拓扑情况。
func (c *ElbClient) ShowLoadbalancersStatus(request *model.ShowLoadbalancersStatusRequest) (*model.ShowLoadbalancersStatusResponse, error) {
	requestDef := GenReqDefForShowLoadbalancersStatus()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowLoadbalancersStatusResponse), nil
	}
}

//根据指定ID查询后端云服务器详情。
func (c *ElbClient) ShowMember(request *model.ShowMemberRequest) (*model.ShowMemberResponse, error) {
	requestDef := GenReqDefForShowMember()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowMemberResponse), nil
	}
}

//根据指定ID查询后端云服务器组详情。
func (c *ElbClient) ShowPool(request *model.ShowPoolRequest) (*model.ShowPoolResponse, error) {
	requestDef := GenReqDefForShowPool()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowPoolResponse), nil
	}
}

//根据指定ID查询白名单详情。
func (c *ElbClient) ShowWhitelist(request *model.ShowWhitelistRequest) (*model.ShowWhitelistResponse, error) {
	requestDef := GenReqDefForShowWhitelist()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowWhitelistResponse), nil
	}
}

//更新健康检查
func (c *ElbClient) UpdateHealthmonitor(request *model.UpdateHealthmonitorRequest) (*model.UpdateHealthmonitorResponse, error) {
	requestDef := GenReqDefForUpdateHealthmonitor()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateHealthmonitorResponse), nil
	}
}

//更新转发策略
func (c *ElbClient) UpdateL7policies(request *model.UpdateL7policiesRequest) (*model.UpdateL7policiesResponse, error) {
	requestDef := GenReqDefForUpdateL7policies()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateL7policiesResponse), nil
	}
}

//更新指定的转发规则
func (c *ElbClient) UpdateL7rule(request *model.UpdateL7ruleRequest) (*model.UpdateL7ruleResponse, error) {
	requestDef := GenReqDefForUpdateL7rule()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateL7ruleResponse), nil
	}
}

//更新监听器。
func (c *ElbClient) UpdateListener(request *model.UpdateListenerRequest) (*model.UpdateListenerResponse, error) {
	requestDef := GenReqDefForUpdateListener()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateListenerResponse), nil
	}
}

//更新负载均衡器。
func (c *ElbClient) UpdateLoadbalancer(request *model.UpdateLoadbalancerRequest) (*model.UpdateLoadbalancerResponse, error) {
	requestDef := GenReqDefForUpdateLoadbalancer()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateLoadbalancerResponse), nil
	}
}

//更新后端云服务器
func (c *ElbClient) UpdateMember(request *model.UpdateMemberRequest) (*model.UpdateMemberResponse, error) {
	requestDef := GenReqDefForUpdateMember()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateMemberResponse), nil
	}
}

//更新后端云服务器组。
func (c *ElbClient) UpdatePool(request *model.UpdatePoolRequest) (*model.UpdatePoolResponse, error) {
	requestDef := GenReqDefForUpdatePool()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdatePoolResponse), nil
	}
}

//更新白名单。可以打开或关闭白名单，或更新访问控制的IP。
func (c *ElbClient) UpdateWhitelist(request *model.UpdateWhitelistRequest) (*model.UpdateWhitelistResponse, error) {
	requestDef := GenReqDefForUpdateWhitelist()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateWhitelistResponse), nil
	}
}

//创建SSL证书。将监听器和SSL证书绑定后，可以通过负载均衡器实现服务端认证，后端服务器只要提供HTTP服务就能实现安全可靠的连接。
func (c *ElbClient) CreateCertificate(request *model.CreateCertificateRequest) (*model.CreateCertificateResponse, error) {
	requestDef := GenReqDefForCreateCertificate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateCertificateResponse), nil
	}
}

//删除指定的SSL证书
func (c *ElbClient) DeleteCertificate(request *model.DeleteCertificateRequest) (*model.DeleteCertificateResponse, error) {
	requestDef := GenReqDefForDeleteCertificate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteCertificateResponse), nil
	}
}

//查询SSL证书。支持过滤查询和分页查询。
func (c *ElbClient) ListCertificates(request *model.ListCertificatesRequest) (*model.ListCertificatesResponse, error) {
	requestDef := GenReqDefForListCertificates()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListCertificatesResponse), nil
	}
}

//查询指定SSL证书的详情信息。
func (c *ElbClient) ShowCertificate(request *model.ShowCertificateRequest) (*model.ShowCertificateResponse, error) {
	requestDef := GenReqDefForShowCertificate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowCertificateResponse), nil
	}
}

//更新指定的SSL证书
func (c *ElbClient) UpdateCertificate(request *model.UpdateCertificateRequest) (*model.UpdateCertificateResponse, error) {
	requestDef := GenReqDefForUpdateCertificate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateCertificateResponse), nil
	}
}
