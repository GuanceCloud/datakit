package v3

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eip/v3/model"
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

//绑定弹性公网IP
func (c *EipClient) AssociatePublicips(request *model.AssociatePublicipsRequest) (*model.AssociatePublicipsResponse, error) {
	requestDef := GenReqDefForAssociatePublicips()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.AssociatePublicipsResponse), nil
	}
}

//解绑弹性公网IP
func (c *EipClient) DisassociatePublicips(request *model.DisassociatePublicipsRequest) (*model.DisassociatePublicipsResponse, error) {
	requestDef := GenReqDefForDisassociatePublicips()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DisassociatePublicipsResponse), nil
	}
}

//查询弹性公网IP列表信息
func (c *EipClient) ListPublicips(request *model.ListPublicipsRequest) (*model.ListPublicipsResponse, error) {
	requestDef := GenReqDefForListPublicips()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListPublicipsResponse), nil
	}
}

//查询弹性公网IP详情
func (c *EipClient) ShowPublicip(request *model.ShowPublicipRequest) (*model.ShowPublicipResponse, error) {
	requestDef := GenReqDefForShowPublicip()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowPublicipResponse), nil
	}
}
