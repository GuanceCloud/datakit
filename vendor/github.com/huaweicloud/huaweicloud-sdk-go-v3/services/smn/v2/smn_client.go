package v2

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/smn/v2/model"
)

type SmnClient struct {
	HcClient *http_client.HcHttpClient
}

func NewSmnClient(hcClient *http_client.HcHttpClient) *SmnClient {
	return &SmnClient{HcClient: hcClient}
}

func SmnClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//为指定Topic添加一个订阅者，如果订阅者的状态为未确认，则向订阅者发送一个确认的消息。待订阅者进行ConfirmSubscription确认后，该订阅者才能收到Topic发布的消息。单Topic默认可添加10000个订阅者，高并发场景下，可能会出现订阅者数量超过10000仍添加成功的情况，此为正常现象。接口是幂等的，如果添加已存在的订阅者，则返回成功，且status code为200，否则status code为201。
func (c *SmnClient) AddSubscription(request *model.AddSubscriptionRequest) (*model.AddSubscriptionResponse, error) {
	requestDef := GenReqDefForAddSubscription()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.AddSubscriptionResponse), nil
	}
}

//为指定实例批量添加或删除标签。一个资源上最多有10个标签。 此接口为幂等接口：创建时如果请求体中存在重复key则报错。 创建时，不允许重复key，如果数据库存在就覆盖。 删除时，如果删除的标签不存在，默认处理成功，删除时不对标签字符集范围做校验。
func (c *SmnClient) BatchCreateOrDeleteResourceTags(request *model.BatchCreateOrDeleteResourceTagsRequest) (*model.BatchCreateOrDeleteResourceTagsResponse, error) {
	requestDef := GenReqDefForBatchCreateOrDeleteResourceTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.BatchCreateOrDeleteResourceTagsResponse), nil
	}
}

//删除指定的订阅者。
func (c *SmnClient) CancelSubscription(request *model.CancelSubscriptionRequest) (*model.CancelSubscriptionResponse, error) {
	requestDef := GenReqDefForCancelSubscription()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CancelSubscriptionResponse), nil
	}
}

//创建一个模板，用户可以按照模板去发送消息，这样可以减少请求的数据量。 单用户默认可创建100个消息模板，高并发场景下，可能会出现消息模板数量超过100仍创建成功的情况，此为正常现象。
func (c *SmnClient) CreateMessageTemplate(request *model.CreateMessageTemplateRequest) (*model.CreateMessageTemplateResponse, error) {
	requestDef := GenReqDefForCreateMessageTemplate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateMessageTemplateResponse), nil
	}
}

//一个资源上最多有10个标签。此接口为幂等接口：创建时，如果创建的标签已经存在（key相同），则覆盖。
func (c *SmnClient) CreateResourceTag(request *model.CreateResourceTagRequest) (*model.CreateResourceTagResponse, error) {
	requestDef := GenReqDefForCreateResourceTag()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateResourceTagResponse), nil
	}
}

//创建Topic，单用户默认配额为3000。高并发场景下，可能会出现Topic数量超过3000仍创建成功的情况，此为正常现象。 接口是幂等的，接口调用返回成功时，若已存在同名的Topic，返回的status code为200，否则返回的status code为201
func (c *SmnClient) CreateTopic(request *model.CreateTopicRequest) (*model.CreateTopicResponse, error) {
	requestDef := GenReqDefForCreateTopic()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateTopicResponse), nil
	}
}

//删除消息模板。删除模板之前的消息请求都可以使用该模板发送，删除之后无法再使用该模板发送消息。
func (c *SmnClient) DeleteMessageTemplate(request *model.DeleteMessageTemplateRequest) (*model.DeleteMessageTemplateResponse, error) {
	requestDef := GenReqDefForDeleteMessageTemplate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteMessageTemplateResponse), nil
	}
}

//幂等接口：删除时，不对标签做校验。删除的key不存在报404，key不能为空或者空字符串。
func (c *SmnClient) DeleteResourceTag(request *model.DeleteResourceTagRequest) (*model.DeleteResourceTagResponse, error) {
	requestDef := GenReqDefForDeleteResourceTag()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteResourceTagResponse), nil
	}
}

//删除主题。
func (c *SmnClient) DeleteTopic(request *model.DeleteTopicRequest) (*model.DeleteTopicResponse, error) {
	requestDef := GenReqDefForDeleteTopic()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteTopicResponse), nil
	}
}

//删除指定名称的主题策略。
func (c *SmnClient) DeleteTopicAttributeByName(request *model.DeleteTopicAttributeByNameRequest) (*model.DeleteTopicAttributeByNameResponse, error) {
	requestDef := GenReqDefForDeleteTopicAttributeByName()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteTopicAttributeByNameResponse), nil
	}
}

//删除所有主题策略。
func (c *SmnClient) DeleteTopicAttributes(request *model.DeleteTopicAttributesRequest) (*model.DeleteTopicAttributesResponse, error) {
	requestDef := GenReqDefForDeleteTopicAttributes()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteTopicAttributesResponse), nil
	}
}

//查询模板详情，包括模板内容。
func (c *SmnClient) ListMessageTemplateDetails(request *model.ListMessageTemplateDetailsRequest) (*model.ListMessageTemplateDetailsResponse, error) {
	requestDef := GenReqDefForListMessageTemplateDetails()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListMessageTemplateDetailsResponse), nil
	}
}

//分页查询模板列表，模板列表按照创建时间进行升序排列。分页查询可以指定offset以及limit。如果不存在模板，则返回空列表。额外的查询参数分别有message_template_name和protocol。
func (c *SmnClient) ListMessageTemplates(request *model.ListMessageTemplatesRequest) (*model.ListMessageTemplatesResponse, error) {
	requestDef := GenReqDefForListMessageTemplates()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListMessageTemplatesResponse), nil
	}
}

//查询租户在指定Region和实例类型的所有标签集合。
func (c *SmnClient) ListProjectTags(request *model.ListProjectTagsRequest) (*model.ListProjectTagsResponse, error) {
	requestDef := GenReqDefForListProjectTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListProjectTagsResponse), nil
	}
}

//使用标签过滤实例。
func (c *SmnClient) ListResourceInstances(request *model.ListResourceInstancesRequest) (*model.ListResourceInstancesResponse, error) {
	requestDef := GenReqDefForListResourceInstances()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListResourceInstancesResponse), nil
	}
}

//查询指定实例的标签信息。
func (c *SmnClient) ListResourceTags(request *model.ListResourceTagsRequest) (*model.ListResourceTagsResponse, error) {
	requestDef := GenReqDefForListResourceTags()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListResourceTagsResponse), nil
	}
}

//分页返回请求者的所有的订阅列表，订阅列表按照订阅创建时间进行升序排列。分页查询可以指定offset以及limit。如果订阅者不存在，返回空列表。
func (c *SmnClient) ListSubscriptions(request *model.ListSubscriptionsRequest) (*model.ListSubscriptionsResponse, error) {
	requestDef := GenReqDefForListSubscriptions()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListSubscriptionsResponse), nil
	}
}

//分页获取特定Topic的订阅列表，订阅列表按照订阅创建时间进行升序排列。分页查询可以指定offset以及limit。如果指定Topic不存在订阅者，返回空列表。
func (c *SmnClient) ListSubscriptionsByTopic(request *model.ListSubscriptionsByTopicRequest) (*model.ListSubscriptionsByTopicResponse, error) {
	requestDef := GenReqDefForListSubscriptionsByTopic()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListSubscriptionsByTopicResponse), nil
	}
}

//查询主题的策略信息。
func (c *SmnClient) ListTopicAttributes(request *model.ListTopicAttributesRequest) (*model.ListTopicAttributesResponse, error) {
	requestDef := GenReqDefForListTopicAttributes()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListTopicAttributesResponse), nil
	}
}

//查询Topic的详细信息。
func (c *SmnClient) ListTopicDetails(request *model.ListTopicDetailsRequest) (*model.ListTopicDetailsResponse, error) {
	requestDef := GenReqDefForListTopicDetails()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListTopicDetailsResponse), nil
	}
}

//分页查询Topic列表，Topic列表按照Topic创建时间进行降序排列。分页查询可以指定offset以及limit。如果不存在Topic，则返回空列表。
func (c *SmnClient) ListTopics(request *model.ListTopicsRequest) (*model.ListTopicsResponse, error) {
	requestDef := GenReqDefForListTopics()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListTopicsResponse), nil
	}
}

//查询SMN API V2版本信息。
func (c *SmnClient) ListVersion(request *model.ListVersionRequest) (*model.ListVersionResponse, error) {
	requestDef := GenReqDefForListVersion()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListVersionResponse), nil
	}
}

//查询SMN开放API支持的版本号。
func (c *SmnClient) ListVersions(request *model.ListVersionsRequest) (*model.ListVersionsResponse, error) {
	requestDef := GenReqDefForListVersions()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListVersionsResponse), nil
	}
}

//将消息发送给Topic的所有订阅端点。当返回消息ID时，该消息已被保存并开始尝试将其推送给Topic的订阅者。三种消息发送方式  message  message_structure  message_template_name  只需要设置其中一个，如果同时设置，生效的优先级为 message_structure > message_template_name > message。
func (c *SmnClient) PublishMessage(request *model.PublishMessageRequest) (*model.PublishMessageResponse, error) {
	requestDef := GenReqDefForPublishMessage()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.PublishMessageResponse), nil
	}
}

//修改消息模板的内容。
func (c *SmnClient) UpdateMessageTemplate(request *model.UpdateMessageTemplateRequest) (*model.UpdateMessageTemplateResponse, error) {
	requestDef := GenReqDefForUpdateMessageTemplate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateMessageTemplateResponse), nil
	}
}

//更新显示名。
func (c *SmnClient) UpdateTopic(request *model.UpdateTopicRequest) (*model.UpdateTopicResponse, error) {
	requestDef := GenReqDefForUpdateTopic()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateTopicResponse), nil
	}
}

//更新主题的策略信息。
func (c *SmnClient) UpdateTopicAttribute(request *model.UpdateTopicAttributeRequest) (*model.UpdateTopicAttributeResponse, error) {
	requestDef := GenReqDefForUpdateTopicAttribute()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateTopicAttributeResponse), nil
	}
}

//创建平台应用。
func (c *SmnClient) CreateApplication(request *model.CreateApplicationRequest) (*model.CreateApplicationResponse, error) {
	requestDef := GenReqDefForCreateApplication()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateApplicationResponse), nil
	}
}

//删除平台应用。
func (c *SmnClient) DeleteApplication(request *model.DeleteApplicationRequest) (*model.DeleteApplicationResponse, error) {
	requestDef := GenReqDefForDeleteApplication()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteApplicationResponse), nil
	}
}

//获取应用平台属性。
func (c *SmnClient) ListApplicationAttributes(request *model.ListApplicationAttributesRequest) (*model.ListApplicationAttributesResponse, error) {
	requestDef := GenReqDefForListApplicationAttributes()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListApplicationAttributesResponse), nil
	}
}

//查询应用平台列表。
func (c *SmnClient) ListApplications(request *model.ListApplicationsRequest) (*model.ListApplicationsResponse, error) {
	requestDef := GenReqDefForListApplications()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListApplicationsResponse), nil
	}
}

//将消息直发给endpoint设备。
func (c *SmnClient) PublishAppMessage(request *model.PublishAppMessageRequest) (*model.PublishAppMessageResponse, error) {
	requestDef := GenReqDefForPublishAppMessage()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.PublishAppMessageResponse), nil
	}
}

//更新应用平台。
func (c *SmnClient) UpdateApplication(request *model.UpdateApplicationRequest) (*model.UpdateApplicationResponse, error) {
	requestDef := GenReqDefForUpdateApplication()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateApplicationResponse), nil
	}
}

//创建应用平台的endpoint终端。
func (c *SmnClient) CreateApplicationEndpoint(request *model.CreateApplicationEndpointRequest) (*model.CreateApplicationEndpointResponse, error) {
	requestDef := GenReqDefForCreateApplicationEndpoint()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateApplicationEndpointResponse), nil
	}
}

//删除设备。
func (c *SmnClient) DeleteApplicationEndpoint(request *model.DeleteApplicationEndpointRequest) (*model.DeleteApplicationEndpointResponse, error) {
	requestDef := GenReqDefForDeleteApplicationEndpoint()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteApplicationEndpointResponse), nil
	}
}

//获取endpoint的属性。
func (c *SmnClient) ListApplicationEndpointAttributes(request *model.ListApplicationEndpointAttributesRequest) (*model.ListApplicationEndpointAttributesResponse, error) {
	requestDef := GenReqDefForListApplicationEndpointAttributes()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListApplicationEndpointAttributesResponse), nil
	}
}

//查询平台的endpoint列表。
func (c *SmnClient) ListApplicationEndpoints(request *model.ListApplicationEndpointsRequest) (*model.ListApplicationEndpointsResponse, error) {
	requestDef := GenReqDefForListApplicationEndpoints()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListApplicationEndpointsResponse), nil
	}
}

//更新设备属性。
func (c *SmnClient) UpdateApplicationEndpoint(request *model.UpdateApplicationEndpointRequest) (*model.UpdateApplicationEndpointResponse, error) {
	requestDef := GenReqDefForUpdateApplicationEndpoint()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateApplicationEndpointResponse), nil
	}
}
