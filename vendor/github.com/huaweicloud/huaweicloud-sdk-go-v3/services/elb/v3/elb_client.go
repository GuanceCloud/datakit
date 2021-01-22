package v3

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v3/model"
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

//创建证书。
func (c *ElbClient) CreateCertificate(request *model.CreateCertificateRequest) (*model.CreateCertificateResponse, error) {
	requestDef := GenReqDefForCreateCertificate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateCertificateResponse), nil
	}
}

//创建健康检查。
func (c *ElbClient) CreateHealthMonitor(request *model.CreateHealthMonitorRequest) (*model.CreateHealthMonitorResponse, error) {
	requestDef := GenReqDefForCreateHealthMonitor()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateHealthMonitorResponse), nil
	}
}

//创建转发策略.
func (c *ElbClient) CreateL7Policy(request *model.CreateL7PolicyRequest) (*model.CreateL7PolicyResponse, error) {
	requestDef := GenReqDefForCreateL7Policy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateL7PolicyResponse), nil
	}
}

//创建转发规则。
func (c *ElbClient) CreateL7Rule(request *model.CreateL7RuleRequest) (*model.CreateL7RuleResponse, error) {
	requestDef := GenReqDefForCreateL7Rule()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateL7RuleResponse), nil
	}
}

//ElbV3 创建监听器。
func (c *ElbClient) CreateListener(request *model.CreateListenerRequest) (*model.CreateListenerResponse, error) {
	requestDef := GenReqDefForCreateListener()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateListenerResponse), nil
	}
}

//创建负载均衡器。 1.创建公网负载均衡器的场合，需要传入vpc_id。 2.创建内网负载均衡器的场合，需要传入vip_subnet_cidr_id。 3.创建内网双栈负载均衡器的场合，需要传入ipv6_vip_virsubnet_id。  关联有已有公网ip地址，需要传入publicip_ids 新建公网ip地址，需要传入publicip 包括IPV4私网类型，IPV4公网类型，IPV6私网，IPV6公网
func (c *ElbClient) CreateLoadBalancer(request *model.CreateLoadBalancerRequest) (*model.CreateLoadBalancerResponse, error) {
	requestDef := GenReqDefForCreateLoadBalancer()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateLoadBalancerResponse), nil
	}
}

//创建后端服务器。
func (c *ElbClient) CreateMember(request *model.CreateMemberRequest) (*model.CreateMemberResponse, error) {
	requestDef := GenReqDefForCreateMember()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateMemberResponse), nil
	}
}

//创建后端服务器组。
func (c *ElbClient) CreatePool(request *model.CreatePoolRequest) (*model.CreatePoolResponse, error) {
	requestDef := GenReqDefForCreatePool()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreatePoolResponse), nil
	}
}

//删除SSL证书。
func (c *ElbClient) DeleteCertificate(request *model.DeleteCertificateRequest) (*model.DeleteCertificateResponse, error) {
	requestDef := GenReqDefForDeleteCertificate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteCertificateResponse), nil
	}
}

//删除健康检查。
func (c *ElbClient) DeleteHealthMonitor(request *model.DeleteHealthMonitorRequest) (*model.DeleteHealthMonitorResponse, error) {
	requestDef := GenReqDefForDeleteHealthMonitor()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteHealthMonitorResponse), nil
	}
}

//删除转发策略。
func (c *ElbClient) DeleteL7Policy(request *model.DeleteL7PolicyRequest) (*model.DeleteL7PolicyResponse, error) {
	requestDef := GenReqDefForDeleteL7Policy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteL7PolicyResponse), nil
	}
}

//删除转发规则。
func (c *ElbClient) DeleteL7Rule(request *model.DeleteL7RuleRequest) (*model.DeleteL7RuleResponse, error) {
	requestDef := GenReqDefForDeleteL7Rule()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteL7RuleResponse), nil
	}
}

//删除监听器。
func (c *ElbClient) DeleteListener(request *model.DeleteListenerRequest) (*model.DeleteListenerResponse, error) {
	requestDef := GenReqDefForDeleteListener()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteListenerResponse), nil
	}
}

//删除负载均衡器。
func (c *ElbClient) DeleteLoadBalancer(request *model.DeleteLoadBalancerRequest) (*model.DeleteLoadBalancerResponse, error) {
	requestDef := GenReqDefForDeleteLoadBalancer()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteLoadBalancerResponse), nil
	}
}

//删除后端服务器。
func (c *ElbClient) DeleteMember(request *model.DeleteMemberRequest) (*model.DeleteMemberResponse, error) {
	requestDef := GenReqDefForDeleteMember()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteMemberResponse), nil
	}
}

//删除后端服务器组。
func (c *ElbClient) DeletePool(request *model.DeletePoolRequest) (*model.DeletePoolResponse, error) {
	requestDef := GenReqDefForDeletePool()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeletePoolResponse), nil
	}
}

//返回租户创建LB时可使用的可用区列表情况。  返回的数据类型是可用区集合的列表，比如列表 [ [az1,az2],  [az2, az3] ] ，有两个可用区集合。在创建负载均衡器时，可以选择创建在多个可用区，但所选的多个可用区必须同属于其中一个可用区集合，如可以选择 az2和az3，但不能选择 az1和az3。
func (c *ElbClient) ListAvailabilityZones(request *model.ListAvailabilityZonesRequest) (*model.ListAvailabilityZonesResponse, error) {
	requestDef := GenReqDefForListAvailabilityZones()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListAvailabilityZonesResponse), nil
	}
}

//查询SSL证书列表。
func (c *ElbClient) ListCertificates(request *model.ListCertificatesRequest) (*model.ListCertificatesResponse, error) {
	requestDef := GenReqDefForListCertificates()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListCertificatesResponse), nil
	}
}

//查询所有的规格。
func (c *ElbClient) ListFlavors(request *model.ListFlavorsRequest) (*model.ListFlavorsResponse, error) {
	requestDef := GenReqDefForListFlavors()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListFlavorsResponse), nil
	}
}

//健康检查列表。
func (c *ElbClient) ListHealthMonitors(request *model.ListHealthMonitorsRequest) (*model.ListHealthMonitorsResponse, error) {
	requestDef := GenReqDefForListHealthMonitors()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListHealthMonitorsResponse), nil
	}
}

//查询转发策略列表。
func (c *ElbClient) ListL7Policies(request *model.ListL7PoliciesRequest) (*model.ListL7PoliciesResponse, error) {
	requestDef := GenReqDefForListL7Policies()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListL7PoliciesResponse), nil
	}
}

//查询转发规则列表。
func (c *ElbClient) ListL7Rules(request *model.ListL7RulesRequest) (*model.ListL7RulesResponse, error) {
	requestDef := GenReqDefForListL7Rules()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListL7RulesResponse), nil
	}
}

//查询监听器列表。
func (c *ElbClient) ListListeners(request *model.ListListenersRequest) (*model.ListListenersResponse, error) {
	requestDef := GenReqDefForListListeners()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListListenersResponse), nil
	}
}

//查询负载均衡器列表，支持过滤查询和分页查询
func (c *ElbClient) ListLoadBalancers(request *model.ListLoadBalancersRequest) (*model.ListLoadBalancersResponse, error) {
	requestDef := GenReqDefForListLoadBalancers()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListLoadBalancersResponse), nil
	}
}

//Pool下的后端服务器列表。
func (c *ElbClient) ListMembers(request *model.ListMembersRequest) (*model.ListMembersResponse, error) {
	requestDef := GenReqDefForListMembers()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListMembersResponse), nil
	}
}

//后端服务器组列表。
func (c *ElbClient) ListPools(request *model.ListPoolsRequest) (*model.ListPoolsResponse, error) {
	requestDef := GenReqDefForListPools()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListPoolsResponse), nil
	}
}

//查询SSL证书详情。
func (c *ElbClient) ShowCertificate(request *model.ShowCertificateRequest) (*model.ShowCertificateResponse, error) {
	requestDef := GenReqDefForShowCertificate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowCertificateResponse), nil
	}
}

//查询规格的详情。
func (c *ElbClient) ShowFlavor(request *model.ShowFlavorRequest) (*model.ShowFlavorResponse, error) {
	requestDef := GenReqDefForShowFlavor()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowFlavorResponse), nil
	}
}

//查询健康检查详情。
func (c *ElbClient) ShowHealthMonitor(request *model.ShowHealthMonitorRequest) (*model.ShowHealthMonitorResponse, error) {
	requestDef := GenReqDefForShowHealthMonitor()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowHealthMonitorResponse), nil
	}
}

//查询转发策略详情。
func (c *ElbClient) ShowL7Policy(request *model.ShowL7PolicyRequest) (*model.ShowL7PolicyResponse, error) {
	requestDef := GenReqDefForShowL7Policy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowL7PolicyResponse), nil
	}
}

//查询转发规则详情
func (c *ElbClient) ShowL7Rule(request *model.ShowL7RuleRequest) (*model.ShowL7RuleResponse, error) {
	requestDef := GenReqDefForShowL7Rule()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowL7RuleResponse), nil
	}
}

//监听器详情。
func (c *ElbClient) ShowListener(request *model.ShowListenerRequest) (*model.ShowListenerResponse, error) {
	requestDef := GenReqDefForShowListener()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowListenerResponse), nil
	}
}

//查询负载均衡器详情
func (c *ElbClient) ShowLoadBalancer(request *model.ShowLoadBalancerRequest) (*model.ShowLoadBalancerResponse, error) {
	requestDef := GenReqDefForShowLoadBalancer()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowLoadBalancerResponse), nil
	}
}

//查询负载均衡器状态树，列出负载均衡器关联的子资源的信息
func (c *ElbClient) ShowLoadBalancerStatus(request *model.ShowLoadBalancerStatusRequest) (*model.ShowLoadBalancerStatusResponse, error) {
	requestDef := GenReqDefForShowLoadBalancerStatus()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowLoadBalancerStatusResponse), nil
	}
}

//后端服务器详情
func (c *ElbClient) ShowMember(request *model.ShowMemberRequest) (*model.ShowMemberResponse, error) {
	requestDef := GenReqDefForShowMember()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowMemberResponse), nil
	}
}

//后端服务器组详情。
func (c *ElbClient) ShowPool(request *model.ShowPoolRequest) (*model.ShowPoolResponse, error) {
	requestDef := GenReqDefForShowPool()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowPoolResponse), nil
	}
}

//【不开放】查询特定项目的配额数。
func (c *ElbClient) ShowQuota(request *model.ShowQuotaRequest) (*model.ShowQuotaResponse, error) {
	requestDef := GenReqDefForShowQuota()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowQuotaResponse), nil
	}
}

//【不开放】查询默认配额数。
func (c *ElbClient) ShowQuotaDefaults(request *model.ShowQuotaDefaultsRequest) (*model.ShowQuotaDefaultsResponse, error) {
	requestDef := GenReqDefForShowQuotaDefaults()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowQuotaDefaultsResponse), nil
	}
}

//更新SSL证书。
func (c *ElbClient) UpdateCertificate(request *model.UpdateCertificateRequest) (*model.UpdateCertificateResponse, error) {
	requestDef := GenReqDefForUpdateCertificate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateCertificateResponse), nil
	}
}

//更新健康检查。
func (c *ElbClient) UpdateHealthMonitor(request *model.UpdateHealthMonitorRequest) (*model.UpdateHealthMonitorResponse, error) {
	requestDef := GenReqDefForUpdateHealthMonitor()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateHealthMonitorResponse), nil
	}
}

//更新转发策略。
func (c *ElbClient) UpdateL7Policy(request *model.UpdateL7PolicyRequest) (*model.UpdateL7PolicyResponse, error) {
	requestDef := GenReqDefForUpdateL7Policy()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateL7PolicyResponse), nil
	}
}

//更新转发规则。
func (c *ElbClient) UpdateL7Rule(request *model.UpdateL7RuleRequest) (*model.UpdateL7RuleResponse, error) {
	requestDef := GenReqDefForUpdateL7Rule()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateL7RuleResponse), nil
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
func (c *ElbClient) UpdateLoadBalancer(request *model.UpdateLoadBalancerRequest) (*model.UpdateLoadBalancerResponse, error) {
	requestDef := GenReqDefForUpdateLoadBalancer()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateLoadBalancerResponse), nil
	}
}

//如果member绑定的负载均衡器的provisioning status不是ACTIVE，则不能更新该member。
func (c *ElbClient) UpdateMember(request *model.UpdateMemberRequest) (*model.UpdateMemberResponse, error) {
	requestDef := GenReqDefForUpdateMember()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateMemberResponse), nil
	}
}

//更新后端服务器组。
func (c *ElbClient) UpdatePool(request *model.UpdatePoolRequest) (*model.UpdatePoolResponse, error) {
	requestDef := GenReqDefForUpdatePool()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdatePoolResponse), nil
	}
}

//计算创建一个负载均衡实例和第一个七层监听器需预占用的IP数
func (c *ElbClient) CountPreoccupyIpNum(request *model.CountPreoccupyIpNumRequest) (*model.CountPreoccupyIpNumResponse, error) {
	requestDef := GenReqDefForCountPreoccupyIpNum()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CountPreoccupyIpNumResponse), nil
	}
}

//创建ip地址组
func (c *ElbClient) CreateIpGroup(request *model.CreateIpGroupRequest) (*model.CreateIpGroupResponse, error) {
	requestDef := GenReqDefForCreateIpGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateIpGroupResponse), nil
	}
}

//删除ip地址组。
func (c *ElbClient) DeleteIpGroup(request *model.DeleteIpGroupRequest) (*model.DeleteIpGroupResponse, error) {
	requestDef := GenReqDefForDeleteIpGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteIpGroupResponse), nil
	}
}

//查询IP地址组列表
func (c *ElbClient) ListIpGroups(request *model.ListIpGroupsRequest) (*model.ListIpGroupsResponse, error) {
	requestDef := GenReqDefForListIpGroups()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListIpGroupsResponse), nil
	}
}

//获取ip地址组详情
func (c *ElbClient) ShowIpGroup(request *model.ShowIpGroupRequest) (*model.ShowIpGroupResponse, error) {
	requestDef := GenReqDefForShowIpGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowIpGroupResponse), nil
	}
}

//更新ip地址组，只支持全量更新ip。
func (c *ElbClient) UpdateIpGroup(request *model.UpdateIpGroupRequest) (*model.UpdateIpGroupResponse, error) {
	requestDef := GenReqDefForUpdateIpGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateIpGroupResponse), nil
	}
}
