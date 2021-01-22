package v1

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dis/v1/model"
)

type DisClient struct {
	HcClient *http_client.HcHttpClient
}

func NewDisClient(hcClient *http_client.HcHttpClient) *DisClient {
	return &DisClient{HcClient: hcClient}
}

func DisClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//本接口用于给指定通道添加权限策略。
func (c *DisClient) CreatePoliciesV3(request *model.CreatePoliciesV3Request) (*model.CreatePoliciesV3Response, error) {
	requestDef := GenReqDefForCreatePoliciesV3()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreatePoliciesV3Response), nil
	}
}

//本接口用于创建通道，推荐使用V3版本接口。
func (c *DisClient) CreateStream(request *model.CreateStreamRequest) (*model.CreateStreamResponse, error) {
	requestDef := GenReqDefForCreateStream()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateStreamResponse), nil
	}
}

//本接口用于添加转储任务。
func (c *DisClient) CreateTransferTask(request *model.CreateTransferTaskRequest) (*model.CreateTransferTaskResponse, error) {
	requestDef := GenReqDefForCreateTransferTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateTransferTaskResponse), nil
	}
}

//本接口用于添加转储任务。
func (c *DisClient) CreateTransferTaskV3(request *model.CreateTransferTaskV3Request) (*model.CreateTransferTaskV3Response, error) {
	requestDef := GenReqDefForCreateTransferTaskV3()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateTransferTaskV3Response), nil
	}
}

//本接口用于删除指定通道，推荐使用V3版本接口。
func (c *DisClient) DeleteStream(request *model.DeleteStreamRequest) (*model.DeleteStreamResponse, error) {
	requestDef := GenReqDefForDeleteStream()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteStreamResponse), nil
	}
}

//\"本接口用于删除指定通道。\"
func (c *DisClient) DeleteStreamV3(request *model.DeleteStreamV3Request) (*model.DeleteStreamV3Response, error) {
	requestDef := GenReqDefForDeleteStreamV3()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteStreamV3Response), nil
	}
}

//This API is used to delete a checkpoint.
func (c *DisClient) DeleteTransferTask(request *model.DeleteTransferTaskRequest) (*model.DeleteTransferTaskResponse, error) {
	requestDef := GenReqDefForDeleteTransferTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteTransferTaskResponse), nil
	}
}

//This API is used to delete a checkpoint.
func (c *DisClient) DeleteTransferTaskV3(request *model.DeleteTransferTaskV3Request) (*model.DeleteTransferTaskV3Response, error) {
	requestDef := GenReqDefForDeleteTransferTaskV3()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteTransferTaskV3Response), nil
	}
}

//本接口用于查询指定通道的详情，推荐使用V3版本接口。
func (c *DisClient) DescribeStream(request *model.DescribeStreamRequest) (*model.DescribeStreamResponse, error) {
	requestDef := GenReqDefForDescribeStream()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DescribeStreamResponse), nil
	}
}

//本接口用于获取数据游标。
func (c *DisClient) GetCursor(request *model.GetCursorRequest) (*model.GetCursorResponse, error) {
	requestDef := GenReqDefForGetCursor()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.GetCursorResponse), nil
	}
}

//本接口用于从DIS通道中下载数据。
func (c *DisClient) GetRecords(request *model.GetRecordsRequest) (*model.GetRecordsResponse, error) {
	requestDef := GenReqDefForGetRecords()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.GetRecordsResponse), nil
	}
}

//本接口用于查询指定通道的权限策略列表。
func (c *DisClient) ListPoliciesV3(request *model.ListPoliciesV3Request) (*model.ListPoliciesV3Response, error) {
	requestDef := GenReqDefForListPoliciesV3()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListPoliciesV3Response), nil
	}
}

//本接口用户查询当前租户创建的所有通道，推荐使用V3版本接口。  查询时，需要指定从哪个通道开始返回通道列表和单次请求需要返回的最大数量。
func (c *DisClient) ListStreams(request *model.ListStreamsRequest) (*model.ListStreamsResponse, error) {
	requestDef := GenReqDefForListStreams()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListStreamsResponse), nil
	}
}

//本接口用于查询转储任务列表。
func (c *DisClient) ListTransferTasksV3(request *model.ListTransferTasksV3Request) (*model.ListTransferTasksV3Response, error) {
	requestDef := GenReqDefForListTransferTasksV3()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListTransferTasksV3Response), nil
	}
}

//本接口用于上传数据到DIS通道中。
func (c *DisClient) PutRecords(request *model.PutRecordsRequest) (*model.PutRecordsResponse, error) {
	requestDef := GenReqDefForPutRecords()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.PutRecordsResponse), nil
	}
}

//本接口用于变更指定通道中的分区数量，推荐使用V3版本接口。
func (c *DisClient) UpdatePartitionCount(request *model.UpdatePartitionCountRequest) (*model.UpdatePartitionCountResponse, error) {
	requestDef := GenReqDefForUpdatePartitionCount()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdatePartitionCountResponse), nil
	}
}

//本接口用于更新指定通道的通道信息。
func (c *DisClient) UpdateStreamV3(request *model.UpdateStreamV3Request) (*model.UpdateStreamV3Response, error) {
	requestDef := GenReqDefForUpdateStreamV3()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateStreamV3Response), nil
	}
}

//本接口用于创建消费APP。
func (c *DisClient) CreateApp(request *model.CreateAppRequest) (*model.CreateAppResponse, error) {
	requestDef := GenReqDefForCreateApp()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateAppResponse), nil
	}
}

//本接口用于创建消费APP。
func (c *DisClient) CreateAppV3(request *model.CreateAppV3Request) (*model.CreateAppV3Response, error) {
	requestDef := GenReqDefForCreateAppV3()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateAppV3Response), nil
	}
}

//本接口用于删除APP。
func (c *DisClient) DeleteApp(request *model.DeleteAppRequest) (*model.DeleteAppResponse, error) {
	requestDef := GenReqDefForDeleteApp()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteAppResponse), nil
	}
}

//本接口用于查询APP详情。
func (c *DisClient) DescribeApp(request *model.DescribeAppRequest) (*model.DescribeAppResponse, error) {
	requestDef := GenReqDefForDescribeApp()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DescribeAppResponse), nil
	}
}

//本接口用于查询APP列表。
func (c *DisClient) ListApp(request *model.ListAppRequest) (*model.ListAppResponse, error) {
	requestDef := GenReqDefForListApp()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListAppResponse), nil
	}
}

//本接口用于查询APP列表。
func (c *DisClient) ListAppV3(request *model.ListAppV3Request) (*model.ListAppV3Response, error) {
	requestDef := GenReqDefForListAppV3()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListAppV3Response), nil
	}
}

//本接口用于新增Checkpoint。
func (c *DisClient) CommitCheckpoint(request *model.CommitCheckpointRequest) (*model.CommitCheckpointResponse, error) {
	requestDef := GenReqDefForCommitCheckpoint()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CommitCheckpointResponse), nil
	}
}

//本接口用于删除Checkpoint。
func (c *DisClient) DeleteCheckpoint(request *model.DeleteCheckpointRequest) (*model.DeleteCheckpointResponse, error) {
	requestDef := GenReqDefForDeleteCheckpoint()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteCheckpointResponse), nil
	}
}

//本接口用于查询Checkpoint详情。
func (c *DisClient) GetCheckpoint(request *model.GetCheckpointRequest) (*model.GetCheckpointResponse, error) {
	requestDef := GenReqDefForGetCheckpoint()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.GetCheckpointResponse), nil
	}
}
