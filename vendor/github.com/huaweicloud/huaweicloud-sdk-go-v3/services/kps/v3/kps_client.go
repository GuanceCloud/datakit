package v3

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/kps/v3/model"
)

type KpsClient struct {
	HcClient *http_client.HcHttpClient
}

func NewKpsClient(hcClient *http_client.HcHttpClient) *KpsClient {
	return &KpsClient{HcClient: hcClient}
}

func KpsClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//给指定的虚拟机绑定（替换或重置，替换需提供虚拟机已配置的SSH密钥对私钥；重置不需要提供虚拟机的SSH密钥对私钥）新的SSH密钥对。
func (c *KpsClient) AssociateKeypair(request *model.AssociateKeypairRequest) (*model.AssociateKeypairResponse, error) {
	requestDef := GenReqDefForAssociateKeypair()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.AssociateKeypairResponse), nil
	}
}

//创建和导入SSH密钥对
func (c *KpsClient) CreateKeypair(request *model.CreateKeypairRequest) (*model.CreateKeypairResponse, error) {
	requestDef := GenReqDefForCreateKeypair()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateKeypairResponse), nil
	}
}

//删除操作失败的任务信息。
func (c *KpsClient) DeleteAllFailedTask(request *model.DeleteAllFailedTaskRequest) (*model.DeleteAllFailedTaskResponse, error) {
	requestDef := GenReqDefForDeleteAllFailedTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteAllFailedTaskResponse), nil
	}
}

//删除失败的任务。
func (c *KpsClient) DeleteFailedTask(request *model.DeleteFailedTaskRequest) (*model.DeleteFailedTaskResponse, error) {
	requestDef := GenReqDefForDeleteFailedTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteFailedTaskResponse), nil
	}
}

//删除SSH密钥对。
func (c *KpsClient) DeleteKeypair(request *model.DeleteKeypairRequest) (*model.DeleteKeypairResponse, error) {
	requestDef := GenReqDefForDeleteKeypair()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteKeypairResponse), nil
	}
}

//给指定的虚拟机解除绑定SSH密钥对并恢复SSH密码登录。
func (c *KpsClient) DisassociateKeypair(request *model.DisassociateKeypairRequest) (*model.DisassociateKeypairResponse, error) {
	requestDef := GenReqDefForDisassociateKeypair()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DisassociateKeypairResponse), nil
	}
}

//查询绑定、解绑等操作失败的任务信息。
func (c *KpsClient) ListFailedTask(request *model.ListFailedTaskRequest) (*model.ListFailedTaskResponse, error) {
	requestDef := GenReqDefForListFailedTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListFailedTaskResponse), nil
	}
}

//查询SSH密钥对详细信息
func (c *KpsClient) ListKeypairDetail(request *model.ListKeypairDetailRequest) (*model.ListKeypairDetailResponse, error) {
	requestDef := GenReqDefForListKeypairDetail()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListKeypairDetailResponse), nil
	}
}

//根据SSH密钥对接口返回的task_id，查询SSH密钥对当前任务的执行状态。
func (c *KpsClient) ListKeypairTask(request *model.ListKeypairTaskRequest) (*model.ListKeypairTaskResponse, error) {
	requestDef := GenReqDefForListKeypairTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListKeypairTaskResponse), nil
	}
}

//查询SSH密钥对列表
func (c *KpsClient) ListKeypairs(request *model.ListKeypairsRequest) (*model.ListKeypairsResponse, error) {
	requestDef := GenReqDefForListKeypairs()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListKeypairsResponse), nil
	}
}

//查询正在处理的任务信息。
func (c *KpsClient) ListRunningTask(request *model.ListRunningTaskRequest) (*model.ListRunningTaskResponse, error) {
	requestDef := GenReqDefForListRunningTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListRunningTaskResponse), nil
	}
}

//更新SSH密钥对描述。
func (c *KpsClient) UpdateKeypairDescription(request *model.UpdateKeypairDescriptionRequest) (*model.UpdateKeypairDescriptionResponse, error) {
	requestDef := GenReqDefForUpdateKeypairDescription()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateKeypairDescriptionResponse), nil
	}
}
