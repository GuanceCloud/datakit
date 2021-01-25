package v2

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eip/v2/model"
)

type EipClient struct {
	HcClient *http_client.HcHttpClient
}

func NewEipClient(hcClient *http_client.HcHttpClient) *EipClient {
	return &EipClient{HcClient: hcClient}
}

func EipClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//共享带宽插入弹性公网IP。
func (c *EipClient) AddPublicipsIntoSharedBandwidth(request *model.AddPublicipsIntoSharedBandwidthRequest) (*model.AddPublicipsIntoSharedBandwidthResponse, error) {
	requestDef := GenReqDefForAddPublicipsIntoSharedBandwidth()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.AddPublicipsIntoSharedBandwidthResponse), nil
	}
}

//批量创建共享带宽。
func (c *EipClient) BatchCreateSharedBandwidths(request *model.BatchCreateSharedBandwidthsRequest) (*model.BatchCreateSharedBandwidthsResponse, error) {
	requestDef := GenReqDefForBatchCreateSharedBandwidths()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.BatchCreateSharedBandwidthsResponse), nil
	}
}

//创建共享带宽。
func (c *EipClient) CreateSharedBandwidth(request *model.CreateSharedBandwidthRequest) (*model.CreateSharedBandwidthResponse, error) {
	requestDef := GenReqDefForCreateSharedBandwidth()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateSharedBandwidthResponse), nil
	}
}

//删除共享带宽。
func (c *EipClient) DeleteSharedBandwidth(request *model.DeleteSharedBandwidthRequest) (*model.DeleteSharedBandwidthResponse, error) {
	requestDef := GenReqDefForDeleteSharedBandwidth()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteSharedBandwidthResponse), nil
	}
}

//查询带宽列表。
func (c *EipClient) ListBandwidths(request *model.ListBandwidthsRequest) (*model.ListBandwidthsResponse, error) {
	requestDef := GenReqDefForListBandwidths()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListBandwidthsResponse), nil
	}
}

//查询配额
func (c *EipClient) ListQuotas(request *model.ListQuotasRequest) (*model.ListQuotasResponse, error) {
	requestDef := GenReqDefForListQuotas()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListQuotasResponse), nil
	}
}

//共享带宽移除弹性公网IP。
func (c *EipClient) RemovePublicipsFromSharedBandwidth(request *model.RemovePublicipsFromSharedBandwidthRequest) (*model.RemovePublicipsFromSharedBandwidthResponse, error) {
	requestDef := GenReqDefForRemovePublicipsFromSharedBandwidth()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.RemovePublicipsFromSharedBandwidthResponse), nil
	}
}

//查询带宽
func (c *EipClient) ShowBandwidth(request *model.ShowBandwidthRequest) (*model.ShowBandwidthResponse, error) {
	requestDef := GenReqDefForShowBandwidth()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowBandwidthResponse), nil
	}
}

//更新带宽。
func (c *EipClient) UpdateBandwidth(request *model.UpdateBandwidthRequest) (*model.UpdateBandwidthResponse, error) {
	requestDef := GenReqDefForUpdateBandwidth()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateBandwidthResponse), nil
	}
}

//更新带宽。
func (c *EipClient) UpdatePrePaidBandwidth(request *model.UpdatePrePaidBandwidthRequest) (*model.UpdatePrePaidBandwidthResponse, error) {
	requestDef := GenReqDefForUpdatePrePaidBandwidth()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdatePrePaidBandwidthResponse), nil
	}
}

//为指定的弹性公网IP资源实例批量添加标签。
func (c *EipClient) BatchCreatePublicipTags(request *model.BatchCreatePublicipTagsRequest) (*model.BatchCreatePublicipTagsResponse, error) {
	requestDef := GenReqDefForBatchCreatePublicipTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.BatchCreatePublicipTagsResponse), nil
	}
}

//为指定的弹性公网IP资源实例批量删除标签。
func (c *EipClient) BatchDeletePublicipTags(request *model.BatchDeletePublicipTagsRequest) (*model.BatchDeletePublicipTagsResponse, error) {
	requestDef := GenReqDefForBatchDeletePublicipTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.BatchDeletePublicipTagsResponse), nil
	}
}

//申请包年包月的弹性公网IP。
func (c *EipClient) CreatePrePaidPublicip(request *model.CreatePrePaidPublicipRequest) (*model.CreatePrePaidPublicipResponse, error) {
	requestDef := GenReqDefForCreatePrePaidPublicip()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreatePrePaidPublicipResponse), nil
	}
}

//申请弹性公网IP，支持IPv4和IPv6。  弹性公网IP（Elastic IP）提供独立的公网IP资源，包括公网IP地址与公网出口带宽服务。可以与弹性云服务器、裸金属服务器、虚拟IP、弹性负载均衡、NAT网关等资源灵活地绑定及解绑。拥有多种灵活的计费方式，可以满足各种业务场景的需要。
func (c *EipClient) CreatePublicip(request *model.CreatePublicipRequest) (*model.CreatePublicipResponse, error) {
	requestDef := GenReqDefForCreatePublicip()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreatePublicipResponse), nil
	}
}

//给指定弹性IP资源实例增加标签信息。
func (c *EipClient) CreatePublicipTag(request *model.CreatePublicipTagRequest) (*model.CreatePublicipTagResponse, error) {
	requestDef := GenReqDefForCreatePublicipTag()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreatePublicipTagResponse), nil
	}
}

//删除弹性公网IP,绑定状态eip不允许直接删除。
func (c *EipClient) DeletePublicip(request *model.DeletePublicipRequest) (*model.DeletePublicipResponse, error) {
	requestDef := GenReqDefForDeletePublicip()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeletePublicipResponse), nil
	}
}

//删除指定弹性公网IP的标签信息。其中project_id是项目ID，publicip_id 是要操作的弹性公网IP的id。key是要删除标签的键。
func (c *EipClient) DeletePublicipTag(request *model.DeletePublicipTagRequest) (*model.DeletePublicipTagResponse, error) {
	requestDef := GenReqDefForDeletePublicipTag()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeletePublicipTagResponse), nil
	}
}

//查询租户在指定区域和实例类型的所有标签集合。
func (c *EipClient) ListPublicipTags(request *model.ListPublicipTagsRequest) (*model.ListPublicipTagsResponse, error) {
	requestDef := GenReqDefForListPublicipTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListPublicipTagsResponse), nil
	}
}

//查询弹性公网IP列表
func (c *EipClient) ListPublicips(request *model.ListPublicipsRequest) (*model.ListPublicipsResponse, error) {
	requestDef := GenReqDefForListPublicips()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListPublicipsResponse), nil
	}
}

//使用标签过滤弹性公网IP。
func (c *EipClient) ListPublicipsByTags(request *model.ListPublicipsByTagsRequest) (*model.ListPublicipsByTagsResponse, error) {
	requestDef := GenReqDefForListPublicipsByTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListPublicipsByTagsResponse), nil
	}
}

//查询指定的弹性公网IP。
func (c *EipClient) ShowPublicip(request *model.ShowPublicipRequest) (*model.ShowPublicipResponse, error) {
	requestDef := GenReqDefForShowPublicip()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowPublicipResponse), nil
	}
}

//查询指定弹性IP实例的标签信息。
func (c *EipClient) ShowPublicipTags(request *model.ShowPublicipTagsRequest) (*model.ShowPublicipTagsResponse, error) {
	requestDef := GenReqDefForShowPublicipTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowPublicipTagsResponse), nil
	}
}

//更新弹性公网IP，将弹性公网IP跟一个网卡绑定或者解绑定，转换IP地址版本类型。
func (c *EipClient) UpdatePublicip(request *model.UpdatePublicipRequest) (*model.UpdatePublicipResponse, error) {
	requestDef := GenReqDefForUpdatePublicip()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdatePublicipResponse), nil
	}
}

//创建浮动IP的外部网络UUID，请使用GET /v2.0/networks?router:external=True或neutron net-external-list方式获取。
func (c *EipClient) NeutronCreateFloatingIp(request *model.NeutronCreateFloatingIpRequest) (*model.NeutronCreateFloatingIpResponse, error) {
	requestDef := GenReqDefForNeutronCreateFloatingIp()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.NeutronCreateFloatingIpResponse), nil
	}
}

//删除指定的浮动IP。
func (c *EipClient) NeutronDeleteFloatingIp(request *model.NeutronDeleteFloatingIpRequest) (*model.NeutronDeleteFloatingIpResponse, error) {
	requestDef := GenReqDefForNeutronDeleteFloatingIp()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.NeutronDeleteFloatingIpResponse), nil
	}
}

//查询提交请求的租户有权限操作的所有浮动IP地址。
func (c *EipClient) NeutronListFloatingIps(request *model.NeutronListFloatingIpsRequest) (*model.NeutronListFloatingIpsResponse, error) {
	requestDef := GenReqDefForNeutronListFloatingIps()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.NeutronListFloatingIpsResponse), nil
	}
}

//查询浮动IP详情，包括浮动IP状态，浮动IP所属路由器ID，浮动IP的外部网络ID等等。
func (c *EipClient) NeutronShowFloatingIp(request *model.NeutronShowFloatingIpRequest) (*model.NeutronShowFloatingIpResponse, error) {
	requestDef := GenReqDefForNeutronShowFloatingIp()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.NeutronShowFloatingIpResponse), nil
	}
}

//更新浮动IP。  更新时需在URL中给出浮动IP地址的ID。  port_id 为空，则表示浮动IP从端口解绑。
func (c *EipClient) NeutronUpdateFloatingIp(request *model.NeutronUpdateFloatingIpRequest) (*model.NeutronUpdateFloatingIpResponse, error) {
	requestDef := GenReqDefForNeutronUpdateFloatingIp()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.NeutronUpdateFloatingIpResponse), nil
	}
}
