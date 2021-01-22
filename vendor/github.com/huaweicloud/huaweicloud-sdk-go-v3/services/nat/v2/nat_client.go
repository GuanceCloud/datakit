package v2

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/nat/v2/model"
)

type NatClient struct {
	HcClient *http_client.HcHttpClient
}

func NewNatClient(hcClient *http_client.HcHttpClient) *NatClient {
	return &NatClient{HcClient: hcClient}
}

func NatClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//批量创建DNAT规则。
func (c *NatClient) BatchCreateNatGatewayDnatRules(request *model.BatchCreateNatGatewayDnatRulesRequest) (*model.BatchCreateNatGatewayDnatRulesResponse, error) {
	requestDef := GenReqDefForBatchCreateNatGatewayDnatRules()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.BatchCreateNatGatewayDnatRulesResponse), nil
	}
}

//创建DNAT规则。
func (c *NatClient) CreateNatGatewayDnatRule(request *model.CreateNatGatewayDnatRuleRequest) (*model.CreateNatGatewayDnatRuleResponse, error) {
	requestDef := GenReqDefForCreateNatGatewayDnatRule()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateNatGatewayDnatRuleResponse), nil
	}
}

//删除指定的DNAT规则。
func (c *NatClient) DeleteNatGatewayDnatRule(request *model.DeleteNatGatewayDnatRuleRequest) (*model.DeleteNatGatewayDnatRuleResponse, error) {
	requestDef := GenReqDefForDeleteNatGatewayDnatRule()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteNatGatewayDnatRuleResponse), nil
	}
}

//查询DNAT规则列表。
func (c *NatClient) ListNatGatewayDnatRules(request *model.ListNatGatewayDnatRulesRequest) (*model.ListNatGatewayDnatRulesResponse, error) {
	requestDef := GenReqDefForListNatGatewayDnatRules()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListNatGatewayDnatRulesResponse), nil
	}
}

//查询指定的DNAT规则详情。
func (c *NatClient) ShowNatGatewayDnatRule(request *model.ShowNatGatewayDnatRuleRequest) (*model.ShowNatGatewayDnatRuleResponse, error) {
	requestDef := GenReqDefForShowNatGatewayDnatRule()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowNatGatewayDnatRuleResponse), nil
	}
}

//更新指定的DNAT规则。
func (c *NatClient) UpdateNatGatewayDnatRule(request *model.UpdateNatGatewayDnatRuleRequest) (*model.UpdateNatGatewayDnatRuleResponse, error) {
	requestDef := GenReqDefForUpdateNatGatewayDnatRule()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateNatGatewayDnatRuleResponse), nil
	}
}

//创建公网NAT网关实例。
func (c *NatClient) CreateNatGateway(request *model.CreateNatGatewayRequest) (*model.CreateNatGatewayResponse, error) {
	requestDef := GenReqDefForCreateNatGateway()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateNatGatewayResponse), nil
	}
}

//删除公网NAT网关实例。
func (c *NatClient) DeleteNatGateway(request *model.DeleteNatGatewayRequest) (*model.DeleteNatGatewayResponse, error) {
	requestDef := GenReqDefForDeleteNatGateway()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteNatGatewayResponse), nil
	}
}

//查询公网NAT网关实例列表。
func (c *NatClient) ListNatGateways(request *model.ListNatGatewaysRequest) (*model.ListNatGatewaysResponse, error) {
	requestDef := GenReqDefForListNatGateways()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListNatGatewaysResponse), nil
	}
}

//查询指定的公网NAT网关实例详情。
func (c *NatClient) ShowNatGateway(request *model.ShowNatGatewayRequest) (*model.ShowNatGatewayResponse, error) {
	requestDef := GenReqDefForShowNatGateway()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowNatGatewayResponse), nil
	}
}

//更新公网NAT网关实例。
func (c *NatClient) UpdateNatGateway(request *model.UpdateNatGatewayRequest) (*model.UpdateNatGatewayResponse, error) {
	requestDef := GenReqDefForUpdateNatGateway()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateNatGatewayResponse), nil
	}
}

//创建SNAT规则。
func (c *NatClient) CreateNatGatewaySnatRule(request *model.CreateNatGatewaySnatRuleRequest) (*model.CreateNatGatewaySnatRuleResponse, error) {
	requestDef := GenReqDefForCreateNatGatewaySnatRule()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateNatGatewaySnatRuleResponse), nil
	}
}

//删除指定的SNAT规则。
func (c *NatClient) DeleteNatGatewaySnatRule(request *model.DeleteNatGatewaySnatRuleRequest) (*model.DeleteNatGatewaySnatRuleResponse, error) {
	requestDef := GenReqDefForDeleteNatGatewaySnatRule()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteNatGatewaySnatRuleResponse), nil
	}
}

//查询SNAT规则列表。
func (c *NatClient) ListNatGatewaySnatRules(request *model.ListNatGatewaySnatRulesRequest) (*model.ListNatGatewaySnatRulesResponse, error) {
	requestDef := GenReqDefForListNatGatewaySnatRules()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListNatGatewaySnatRulesResponse), nil
	}
}

//查询指定的SNAT规则详情。
func (c *NatClient) ShowNatGatewaySnatRule(request *model.ShowNatGatewaySnatRuleRequest) (*model.ShowNatGatewaySnatRuleResponse, error) {
	requestDef := GenReqDefForShowNatGatewaySnatRule()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowNatGatewaySnatRuleResponse), nil
	}
}

//更新指定的SNAT规则。
func (c *NatClient) UpdateNatGatewaySnatRule(request *model.UpdateNatGatewaySnatRuleRequest) (*model.UpdateNatGatewaySnatRuleResponse, error) {
	requestDef := GenReqDefForUpdateNatGatewaySnatRule()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateNatGatewaySnatRuleResponse), nil
	}
}
