package v2

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dms/v2/model"
)

type DmsClient struct {
	HcClient *http_client.HcHttpClient
}

func NewDmsClient(hcClient *http_client.HcHttpClient) *DmsClient {
	return &DmsClient{HcClient: hcClient}
}

func DmsClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//批量添加或删除队列标签。
func (c *DmsClient) BatchCreateOrDeleteQueueTag(request *model.BatchCreateOrDeleteQueueTagRequest) (*model.BatchCreateOrDeleteQueueTagResponse, error) {
	requestDef := GenReqDefForBatchCreateOrDeleteQueueTag()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.BatchCreateOrDeleteQueueTagResponse), nil
	}
}

//确认已经消费指定消息。  在消费者消费消息期间，消息仍然停留在队列中，但消息从被消费开始的30秒内不能被该消费组再次消费，若在这30秒内没有被消费者确认消费，则DMS认为消息未消费成功，将可以被继续消费。  如果消息被确认消费成功，消息将不能被该消费组再次消费，但是消息仍然保持在队列中，并且可以被其它消费组消费，消息在队列中的保留时间默认为72小时（除非队列被删除），72小时后会被删除。  消息批量消费确认时，必须严格按照消息消费的顺序提交确认，DMS按顺序判定消息是否消费成功，如果某条消息未确认或消费失败，则不再继续检测，默认后续消息全部消费失败。建议当对某一条消息处理失败时，不再需要继续处理本批消息中的后续消息，直接对已正确处理的消息进行确认。  确认消费失败后，可以再次重新消费和确认。当开启死信时，消息进行多次重复消费仍然失败后，DMS会将该条消息转存到死信队列中，有效期为72小时，用户可以根据需要对死信消息进行重新消费。
func (c *DmsClient) ConfirmConsumptionMessages(request *model.ConfirmConsumptionMessagesRequest) (*model.ConfirmConsumptionMessagesResponse, error) {
	requestDef := GenReqDefForConfirmConsumptionMessages()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ConfirmConsumptionMessagesResponse), nil
	}
}

//确认已经消费指定的死信消息。  在消费者消费死信消息期间，死信消息仍然停留在队列中，但死信消息从被消费开始的30秒内不能被该消费组再次消费，若在这30秒内没有被消费者确认消费，则DMS认为死信消息未消费成功，将可以被继续消费。  如果死信消息被确认消费成功，该死信消息将不能被该消费组再次消费，死信消息的保留时间为72小时（除非消费组被删除），72小时后会被删除。  消息批量消费确认时，必须严格按照消息消费的顺序提交确认，DMS按顺序判定消息是否消费成功，如果某条消息未确认或消费失败，则不再继续检测，默认后续消息全部消费失败。建议当对某一条消息处理失败时，不再需要继续处理本批消息中的后续消息，直接对已正确处理的消息进行确认。  仅NORMAL队列和FIFO队列可以开启死信消息，因为只有NORMAL队列和FIFO队列可消费死信消息。
func (c *DmsClient) ConfirmDeadLettersMessages(request *model.ConfirmDeadLettersMessagesRequest) (*model.ConfirmDeadLettersMessagesResponse, error) {
	requestDef := GenReqDefForConfirmDeadLettersMessages()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ConfirmDeadLettersMessagesResponse), nil
	}
}

//消费指定消费组产生的死信消息。可同时消费多条消息，每次消费的消息负载不超过512KB。  仅NORMAL队列和FIFO队列可以开启死信消息，因为只有NORMAL队列和FIFO队列可消费死信消息。
func (c *DmsClient) ConsumeDeadlettersMessage(request *model.ConsumeDeadlettersMessageRequest) (*model.ConsumeDeadlettersMessageResponse, error) {
	requestDef := GenReqDefForConsumeDeadlettersMessage()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ConsumeDeadlettersMessageResponse), nil
	}
}

//消费指定队列中的消息。可同时消费多条消息，每次消费的消息负载不超过512KB。  当队列中消息较少时，单次消费返回的消息数量可能会少于指定条数，但多次消费最终可获取全部消息，当队列为空时，返回为空。  每个消费组只支持一种Label规则，如果第二次消费更换了Label规则，则消费失败。
func (c *DmsClient) ConsumeMessages(request *model.ConsumeMessagesRequest) (*model.ConsumeMessagesResponse, error) {
	requestDef := GenReqDefForConsumeMessages()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ConsumeMessagesResponse), nil
	}
}

//创建消费组。  可同时为指定队列创建多个消费组。  > 创建消费组后系统内部完成初始化需要1-3秒，如果创建消费组后立即进行操作，可能会导致消费失败。建议3秒后再操作该队列。
func (c *DmsClient) CreateConsumerGroup(request *model.CreateConsumerGroupRequest) (*model.CreateConsumerGroupResponse, error) {
	requestDef := GenReqDefForCreateConsumerGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateConsumerGroupResponse), nil
	}
}

//每个project_id默认只能创建30个队列。 > 创建队列时系统内部完成初始化需要1-3秒，如果创建队列后立即进行操作，可能会导致生产消息、消费消息、查询队列详情、创建消费组和删除队列等操作失败。建议3秒后再操作该队列。
func (c *DmsClient) CreateQueue(request *model.CreateQueueRequest) (*model.CreateQueueResponse, error) {
	requestDef := GenReqDefForCreateQueue()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateQueueResponse), nil
	}
}

//删除指定的队列。
func (c *DmsClient) DeleteQueue(request *model.DeleteQueueRequest) (*model.DeleteQueueResponse, error) {
	requestDef := GenReqDefForDeleteQueue()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteQueueResponse), nil
	}
}

//删除指定消费组。
func (c *DmsClient) DeleteSpecifiedConsumerGroup(request *model.DeleteSpecifiedConsumerGroupRequest) (*model.DeleteSpecifiedConsumerGroupResponse, error) {
	requestDef := GenReqDefForDeleteSpecifiedConsumerGroup()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteSpecifiedConsumerGroupResponse), nil
	}
}

//获取指定队列的所有消费组。
func (c *DmsClient) ListConsumerGroups(request *model.ListConsumerGroupsRequest) (*model.ListConsumerGroupsResponse, error) {
	requestDef := GenReqDefForListConsumerGroups()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListConsumerGroupsResponse), nil
	}
}

//查看所有队列。
func (c *DmsClient) ListQueues(request *model.ListQueuesRequest) (*model.ListQueuesResponse, error) {
	requestDef := GenReqDefForListQueues()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListQueuesResponse), nil
	}
}

//向指定队列发送消息，可同时发送多条消息。  - 每次最多发送10条。 - 每次发送的消息总负载不超过512KB。 - Kafka队列的消息保存时间在创建队列时可以设置，可设置范围为1~72小时。其他队列的消息最大保存时长为72小时。
func (c *DmsClient) SendMessages(request *model.SendMessagesRequest) (*model.SendMessagesResponse, error) {
	requestDef := GenReqDefForSendMessages()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.SendMessagesResponse), nil
	}
}

//查询项目标签。
func (c *DmsClient) ShowProjectTags(request *model.ShowProjectTagsRequest) (*model.ShowProjectTagsResponse, error) {
	requestDef := GenReqDefForShowProjectTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowProjectTagsResponse), nil
	}
}

//查看指定的队列。
func (c *DmsClient) ShowQueue(request *model.ShowQueueRequest) (*model.ShowQueueResponse, error) {
	requestDef := GenReqDefForShowQueue()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowQueueResponse), nil
	}
}

//查询队列标签。
func (c *DmsClient) ShowQueueTags(request *model.ShowQueueTagsRequest) (*model.ShowQueueTagsResponse, error) {
	requestDef := GenReqDefForShowQueueTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowQueueTagsResponse), nil
	}
}

//查看当前项目的配额。
func (c *DmsClient) ShowQuotas(request *model.ShowQuotasRequest) (*model.ShowQuotasResponse, error) {
	requestDef := GenReqDefForShowQuotas()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowQuotasResponse), nil
	}
}
