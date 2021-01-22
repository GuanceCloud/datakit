package v2

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/smn/v2/model"
	"net/http"
)

func GenReqDefForAddSubscription() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v2/{project_id}/notifications/topics/{topic_urn}/subscriptions").
		WithResponse(new(model.AddSubscriptionResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("TopicUrn").
		WithJsonTag("topic_urn").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForBatchCreateOrDeleteResourceTags() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v2/{project_id}/{resource_type}/{resource_id}/tags/action").
		WithResponse(new(model.BatchCreateOrDeleteResourceTagsResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ResourceType").
		WithJsonTag("resource_type").
		WithLocationType(def.Path))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ResourceId").
		WithJsonTag("resource_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForCancelSubscription() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/v2/{project_id}/notifications/subscriptions/{subscription_urn}").
		WithResponse(new(model.CancelSubscriptionResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("SubscriptionUrn").
		WithJsonTag("subscription_urn").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForCreateMessageTemplate() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v2/{project_id}/notifications/message_template").
		WithResponse(new(model.CreateMessageTemplateResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForCreateResourceTag() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v2/{project_id}/{resource_type}/{resource_id}/tags").
		WithResponse(new(model.CreateResourceTagResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ResourceType").
		WithJsonTag("resource_type").
		WithLocationType(def.Path))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ResourceId").
		WithJsonTag("resource_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForCreateTopic() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v2/{project_id}/notifications/topics").
		WithResponse(new(model.CreateTopicResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForDeleteMessageTemplate() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/v2/{project_id}/notifications/message_template/{message_template_id}").
		WithResponse(new(model.DeleteMessageTemplateResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("MessageTemplateId").
		WithJsonTag("message_template_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForDeleteResourceTag() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/v2/{project_id}/{resource_type}/{resource_id}/tags/{key}").
		WithResponse(new(model.DeleteResourceTagResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ResourceType").
		WithJsonTag("resource_type").
		WithLocationType(def.Path))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ResourceId").
		WithJsonTag("resource_id").
		WithLocationType(def.Path))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Key").
		WithJsonTag("key").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForDeleteTopic() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/v2/{project_id}/notifications/topics/{topic_urn}").
		WithResponse(new(model.DeleteTopicResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("TopicUrn").
		WithJsonTag("topic_urn").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForDeleteTopicAttributeByName() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/v2/{project_id}/notifications/topics/{topic_urn}/attributes/{name}").
		WithResponse(new(model.DeleteTopicAttributeByNameResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("TopicUrn").
		WithJsonTag("topic_urn").
		WithLocationType(def.Path))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Name").
		WithJsonTag("name").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForDeleteTopicAttributes() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/v2/{project_id}/notifications/topics/{topic_urn}/attributes").
		WithResponse(new(model.DeleteTopicAttributesResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("TopicUrn").
		WithJsonTag("topic_urn").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListMessageTemplateDetails() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v2/{project_id}/notifications/message_template/{message_template_id}").
		WithResponse(new(model.ListMessageTemplateDetailsResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("MessageTemplateId").
		WithJsonTag("message_template_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListMessageTemplates() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v2/{project_id}/notifications/message_template").
		WithResponse(new(model.ListMessageTemplatesResponse)).
		WithContentType("application/json")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Offset").
		WithJsonTag("offset").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Limit").
		WithJsonTag("limit").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("MessageTemplateName").
		WithJsonTag("message_template_name").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Protocol").
		WithJsonTag("protocol").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListProjectTags() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v2/{project_id}/{resource_type}/tags").
		WithResponse(new(model.ListProjectTagsResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ResourceType").
		WithJsonTag("resource_type").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListResourceInstances() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v2/{project_id}/{resource_type}/resource_instances/action").
		WithResponse(new(model.ListResourceInstancesResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ResourceType").
		WithJsonTag("resource_type").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListResourceTags() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v2/{project_id}/{resource_type}/{resource_id}/tags").
		WithResponse(new(model.ListResourceTagsResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ResourceType").
		WithJsonTag("resource_type").
		WithLocationType(def.Path))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ResourceId").
		WithJsonTag("resource_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListSubscriptions() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v2/{project_id}/notifications/subscriptions").
		WithResponse(new(model.ListSubscriptionsResponse)).
		WithContentType("application/json")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Offset").
		WithJsonTag("offset").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Limit").
		WithJsonTag("limit").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListSubscriptionsByTopic() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v2/{project_id}/notifications/topics/{topic_urn}/subscriptions").
		WithResponse(new(model.ListSubscriptionsByTopicResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("TopicUrn").
		WithJsonTag("topic_urn").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Offset").
		WithJsonTag("offset").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Limit").
		WithJsonTag("limit").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListTopicAttributes() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v2/{project_id}/notifications/topics/{topic_urn}/attributes").
		WithResponse(new(model.ListTopicAttributesResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("TopicUrn").
		WithJsonTag("topic_urn").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Name").
		WithJsonTag("name").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListTopicDetails() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v2/{project_id}/notifications/topics/{topic_urn}").
		WithResponse(new(model.ListTopicDetailsResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("TopicUrn").
		WithJsonTag("topic_urn").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListTopics() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v2/{project_id}/notifications/topics").
		WithResponse(new(model.ListTopicsResponse)).
		WithContentType("application/json")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Offset").
		WithJsonTag("offset").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Limit").
		WithJsonTag("limit").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListVersion() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/{api_version}").
		WithResponse(new(model.ListVersionResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ApiVersion").
		WithJsonTag("api_version").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListVersions() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/").
		WithResponse(new(model.ListVersionsResponse)).
		WithContentType("application/json")

	// request

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForPublishMessage() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v2/{project_id}/notifications/topics/{topic_urn}/publish").
		WithResponse(new(model.PublishMessageResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("TopicUrn").
		WithJsonTag("topic_urn").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForUpdateMessageTemplate() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPut).
		WithPath("/v2/{project_id}/notifications/message_template/{message_template_id}").
		WithResponse(new(model.UpdateMessageTemplateResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("MessageTemplateId").
		WithJsonTag("message_template_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForUpdateTopic() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPut).
		WithPath("/v2/{project_id}/notifications/topics/{topic_urn}").
		WithResponse(new(model.UpdateTopicResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("TopicUrn").
		WithJsonTag("topic_urn").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForUpdateTopicAttribute() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPut).
		WithPath("/v2/{project_id}/notifications/topics/{topic_urn}/attributes/{name}").
		WithResponse(new(model.UpdateTopicAttributeResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("TopicUrn").
		WithJsonTag("topic_urn").
		WithLocationType(def.Path))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Name").
		WithJsonTag("name").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForCreateApplication() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v2/{project_id}/notifications/applications").
		WithResponse(new(model.CreateApplicationResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForDeleteApplication() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/v2/{project_id}/notifications/applications/{application_urn}").
		WithResponse(new(model.DeleteApplicationResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ApplicationUrn").
		WithJsonTag("application_urn").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListApplicationAttributes() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v2/{project_id}/notifications/applications/{application_urn}").
		WithResponse(new(model.ListApplicationAttributesResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ApplicationUrn").
		WithJsonTag("application_urn").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListApplications() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v2/{project_id}/notifications/applications").
		WithResponse(new(model.ListApplicationsResponse)).
		WithContentType("application/json")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Offset").
		WithJsonTag("offset").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Limit").
		WithJsonTag("limit").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Name").
		WithJsonTag("name").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Platform").
		WithJsonTag("platform").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForPublishAppMessage() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v2/{project_id}/notifications/endpoints/{endpoint_urn}/publish").
		WithResponse(new(model.PublishAppMessageResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("EndpointUrn").
		WithJsonTag("endpoint_urn").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForUpdateApplication() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPut).
		WithPath("/v2/{project_id}/notifications/applications/{application_urn}").
		WithResponse(new(model.UpdateApplicationResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ApplicationUrn").
		WithJsonTag("application_urn").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForCreateApplicationEndpoint() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v2/{project_id}/notifications/applications/{application_urn}/endpoints").
		WithResponse(new(model.CreateApplicationEndpointResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ApplicationUrn").
		WithJsonTag("application_urn").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForDeleteApplicationEndpoint() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/v2/{project_id}/notifications/endpoints/{endpoint_urn}").
		WithResponse(new(model.DeleteApplicationEndpointResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("EndpointUrn").
		WithJsonTag("endpoint_urn").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListApplicationEndpointAttributes() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v2/{project_id}/notifications/endpoints/{endpoint_urn}").
		WithResponse(new(model.ListApplicationEndpointAttributesResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("EndpointUrn").
		WithJsonTag("endpoint_urn").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListApplicationEndpoints() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v2/{project_id}/notifications/applications/{application_urn}/endpoints").
		WithResponse(new(model.ListApplicationEndpointsResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ApplicationUrn").
		WithJsonTag("application_urn").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Offset").
		WithJsonTag("offset").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Limit").
		WithJsonTag("limit").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Enabled").
		WithJsonTag("enabled").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Token").
		WithJsonTag("token").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("UserData").
		WithJsonTag("user_data").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForUpdateApplicationEndpoint() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPut).
		WithPath("/v2/{project_id}/notifications/endpoints/{endpoint_urn}").
		WithResponse(new(model.UpdateApplicationEndpointResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("EndpointUrn").
		WithJsonTag("endpoint_urn").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}
