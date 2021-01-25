package v2

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rabbitmq/v2/model"
)

type RabbitMQClient struct {
	HcClient *http_client.HcHttpClient
}

func NewRabbitMQClient(hcClient *http_client.HcHttpClient) *RabbitMQClient {
	return &RabbitMQClient{HcClient: hcClient}
}

func RabbitMQClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//批量添加或删除实例标签。
func (c *RabbitMQClient) BatchCreateOrDeleteInstanceTag(request *model.BatchCreateOrDeleteInstanceTagRequest) (*model.BatchCreateOrDeleteInstanceTagResponse, error) {
	requestDef := GenReqDefForBatchCreateOrDeleteInstanceTag()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.BatchCreateOrDeleteInstanceTagResponse), nil
	}
}

//批量重启或删除实例。  在实例重启过程中，客户端的生产与消费消息等请求会被拒绝。  实例删除后，实例中原有的数据将被删除，且没有备份，请谨慎操作。
func (c *RabbitMQClient) BatchRestartOrDeleteInstances(request *model.BatchRestartOrDeleteInstancesRequest) (*model.BatchRestartOrDeleteInstancesResponse, error) {
	requestDef := GenReqDefForBatchRestartOrDeleteInstances()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.BatchRestartOrDeleteInstancesResponse), nil
	}
}

//创建实例，该接口创建的实例为按需计费的方式。
func (c *RabbitMQClient) CreatePostPaidInstance(request *model.CreatePostPaidInstanceRequest) (*model.CreatePostPaidInstanceResponse, error) {
	requestDef := GenReqDefForCreatePostPaidInstance()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreatePostPaidInstanceResponse), nil
	}
}

//删除后台任务管理中的指定记录。
func (c *RabbitMQClient) DeleteBackgroundTask(request *model.DeleteBackgroundTaskRequest) (*model.DeleteBackgroundTaskResponse, error) {
	requestDef := GenReqDefForDeleteBackgroundTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteBackgroundTaskResponse), nil
	}
}

//删除指定的实例，释放该实例的所有资源。
func (c *RabbitMQClient) DeleteInstance(request *model.DeleteInstanceRequest) (*model.DeleteInstanceResponse, error) {
	requestDef := GenReqDefForDeleteInstance()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteInstanceResponse), nil
	}
}

//在创建实例时，需要配置实例所在的可用区ID，可通过该接口查询可用区的ID。
func (c *RabbitMQClient) ListAvailableZones(request *model.ListAvailableZonesRequest) (*model.ListAvailableZonesResponse, error) {
	requestDef := GenReqDefForListAvailableZones()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListAvailableZonesResponse), nil
	}
}

//查询实例的后台任务列表。
func (c *RabbitMQClient) ListBackgroundTasks(request *model.ListBackgroundTasksRequest) (*model.ListBackgroundTasksResponse, error) {
	requestDef := GenReqDefForListBackgroundTasks()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListBackgroundTasksResponse), nil
	}
}

//查询租户的实例列表，支持按照条件查询。
func (c *RabbitMQClient) ListInstancesDetails(request *model.ListInstancesDetailsRequest) (*model.ListInstancesDetailsResponse, error) {
	requestDef := GenReqDefForListInstancesDetails()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListInstancesDetailsResponse), nil
	}
}

//查询插件列表。
func (c *RabbitMQClient) ListPlugins(request *model.ListPluginsRequest) (*model.ListPluginsResponse, error) {
	requestDef := GenReqDefForListPlugins()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListPluginsResponse), nil
	}
}

//在创建实例时，需要配置订购的产品ID（即product_id），可通过该接口查询产品规格。
func (c *RabbitMQClient) ListProducts(request *model.ListProductsRequest) (*model.ListProductsResponse, error) {
	requestDef := GenReqDefForListProducts()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListProductsResponse), nil
	}
}

//重置密码。
func (c *RabbitMQClient) ResetPassword(request *model.ResetPasswordRequest) (*model.ResetPasswordResponse, error) {
	requestDef := GenReqDefForResetPassword()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ResetPasswordResponse), nil
	}
}

//实例规格变更。
func (c *RabbitMQClient) ResizeInstance(request *model.ResizeInstanceRequest) (*model.ResizeInstanceResponse, error) {
	requestDef := GenReqDefForResizeInstance()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ResizeInstanceResponse), nil
	}
}

//查询后台任务管理中的指定记录。
func (c *RabbitMQClient) ShowBackgroundTask(request *model.ShowBackgroundTaskRequest) (*model.ShowBackgroundTaskResponse, error) {
	requestDef := GenReqDefForShowBackgroundTask()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowBackgroundTaskResponse), nil
	}
}

//查询指定实例的详细信息。
func (c *RabbitMQClient) ShowInstance(request *model.ShowInstanceRequest) (*model.ShowInstanceResponse, error) {
	requestDef := GenReqDefForShowInstance()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowInstanceResponse), nil
	}
}

//查询可扩容规格列表。  RabbtiMQ只支持只增加节点数的扩容方式。
func (c *RabbitMQClient) ShowInstanceExtendProductInfo(request *model.ShowInstanceExtendProductInfoRequest) (*model.ShowInstanceExtendProductInfoResponse, error) {
	requestDef := GenReqDefForShowInstanceExtendProductInfo()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowInstanceExtendProductInfoResponse), nil
	}
}

//查询实例标签。
func (c *RabbitMQClient) ShowInstanceTags(request *model.ShowInstanceTagsRequest) (*model.ShowInstanceTagsResponse, error) {
	requestDef := GenReqDefForShowInstanceTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowInstanceTagsResponse), nil
	}
}

//查询维护时间窗开始时间和结束时间。
func (c *RabbitMQClient) ShowMaintainWindows(request *model.ShowMaintainWindowsRequest) (*model.ShowMaintainWindowsResponse, error) {
	requestDef := GenReqDefForShowMaintainWindows()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowMaintainWindowsResponse), nil
	}
}

//查询项目标签。
func (c *RabbitMQClient) ShowProjectTags(request *model.ShowProjectTagsRequest) (*model.ShowProjectTagsResponse, error) {
	requestDef := GenReqDefForShowProjectTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowProjectTagsResponse), nil
	}
}

//修改实例的名称和描述信息。
func (c *RabbitMQClient) UpdateInstance(request *model.UpdateInstanceRequest) (*model.UpdateInstanceResponse, error) {
	requestDef := GenReqDefForUpdateInstance()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateInstanceResponse), nil
	}
}

//开启或关闭插件。
func (c *RabbitMQClient) UpdatePlugins(request *model.UpdatePluginsRequest) (*model.UpdatePluginsResponse, error) {
	requestDef := GenReqDefForUpdatePlugins()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdatePluginsResponse), nil
	}
}
