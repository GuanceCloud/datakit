package v3

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v3/model"
	"net/http"
)

func GenReqDefForCreateCertificate() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v3/{project_id}/elb/certificates").
		WithResponse(new(model.CreateCertificateResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForCreateHealthMonitor() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v3/{project_id}/elb/healthmonitors").
		WithResponse(new(model.CreateHealthMonitorResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForCreateL7Policy() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v3/{project_id}/elb/l7policies").
		WithResponse(new(model.CreateL7PolicyResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForCreateL7Rule() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v3/{project_id}/elb/l7policies/{l7policy_id}/rules").
		WithResponse(new(model.CreateL7RuleResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("L7policyId").
		WithJsonTag("l7policy_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForCreateListener() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v3/{project_id}/elb/listeners").
		WithResponse(new(model.CreateListenerResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForCreateLoadBalancer() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v3/{project_id}/elb/loadbalancers").
		WithResponse(new(model.CreateLoadBalancerResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForCreateMember() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v3/{project_id}/elb/pools/{pool_id}/members").
		WithResponse(new(model.CreateMemberResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("PoolId").
		WithJsonTag("pool_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForCreatePool() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v3/{project_id}/elb/pools").
		WithResponse(new(model.CreatePoolResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForDeleteCertificate() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/v3/{project_id}/elb/certificates/{certificate_id}").
		WithResponse(new(model.DeleteCertificateResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("CertificateId").
		WithJsonTag("certificate_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForDeleteHealthMonitor() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/v3/{project_id}/elb/healthmonitors/{healthmonitor_id}").
		WithResponse(new(model.DeleteHealthMonitorResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("HealthmonitorId").
		WithJsonTag("healthmonitor_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForDeleteL7Policy() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/v3/{project_id}/elb/l7policies/{l7policy_id}").
		WithResponse(new(model.DeleteL7PolicyResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("L7policyId").
		WithJsonTag("l7policy_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForDeleteL7Rule() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/v3/{project_id}/elb/l7policies/{l7policy_id}/rules/{l7rule_id}").
		WithResponse(new(model.DeleteL7RuleResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("L7policyId").
		WithJsonTag("l7policy_id").
		WithLocationType(def.Path))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("L7ruleId").
		WithJsonTag("l7rule_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForDeleteListener() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/v3/{project_id}/elb/listeners/{listener_id}").
		WithResponse(new(model.DeleteListenerResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ListenerId").
		WithJsonTag("listener_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForDeleteLoadBalancer() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/v3/{project_id}/elb/loadbalancers/{loadbalancer_id}").
		WithResponse(new(model.DeleteLoadBalancerResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("LoadbalancerId").
		WithJsonTag("loadbalancer_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForDeleteMember() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/v3/{project_id}/elb/pools/{pool_id}/members/{member_id}").
		WithResponse(new(model.DeleteMemberResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("MemberId").
		WithJsonTag("member_id").
		WithLocationType(def.Path))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("PoolId").
		WithJsonTag("pool_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForDeletePool() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/v3/{project_id}/elb/pools/{pool_id}").
		WithResponse(new(model.DeletePoolResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("PoolId").
		WithJsonTag("pool_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListAvailabilityZones() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/availability-zones").
		WithResponse(new(model.ListAvailabilityZonesResponse)).
		WithContentType("application/json")

	// request

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListCertificates() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/certificates").
		WithResponse(new(model.ListCertificatesResponse)).
		WithContentType("application/json")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("AdminStateUp").
		WithJsonTag("admin_state_up").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Description").
		WithJsonTag("description").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Domain").
		WithJsonTag("domain").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Id").
		WithJsonTag("id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Limit").
		WithJsonTag("limit").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Marker").
		WithJsonTag("marker").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Name").
		WithJsonTag("name").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("PageReverse").
		WithJsonTag("page_reverse").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Type").
		WithJsonTag("type").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListFlavors() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/flavors").
		WithResponse(new(model.ListFlavorsResponse)).
		WithContentType("application/json")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Id").
		WithJsonTag("id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Limit").
		WithJsonTag("limit").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Marker").
		WithJsonTag("marker").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Name").
		WithJsonTag("name").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("PageReverse").
		WithJsonTag("page_reverse").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Shared").
		WithJsonTag("shared").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Type").
		WithJsonTag("type").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListHealthMonitors() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/healthmonitors").
		WithResponse(new(model.ListHealthMonitorsResponse)).
		WithContentType("application/json")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("AdminStateUp").
		WithJsonTag("admin_state_up").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Delay").
		WithJsonTag("delay").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("DomainName").
		WithJsonTag("domain_name").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("EnterpriseProjectId").
		WithJsonTag("enterprise_project_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ExpectedCodes").
		WithJsonTag("expected_codes").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("HttpMethod").
		WithJsonTag("http_method").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Id").
		WithJsonTag("id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Limit").
		WithJsonTag("limit").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Marker").
		WithJsonTag("marker").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("MaxRetries").
		WithJsonTag("max_retries").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("MaxRetriesDown").
		WithJsonTag("max_retries_down").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("MonitorPort").
		WithJsonTag("monitor_port").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Name").
		WithJsonTag("name").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("PageReverse").
		WithJsonTag("page_reverse").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Timeout").
		WithJsonTag("timeout").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Type").
		WithJsonTag("type").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("UrlPath").
		WithJsonTag("url_path").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListL7Policies() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/l7policies").
		WithResponse(new(model.ListL7PoliciesResponse)).
		WithContentType("application/json")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Action").
		WithJsonTag("action").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("AdminStateUp").
		WithJsonTag("admin_state_up").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Description").
		WithJsonTag("description").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("DisplayAllRules").
		WithJsonTag("display_all_rules").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("EnterpriseProjectId").
		WithJsonTag("enterprise_project_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Id").
		WithJsonTag("id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Limit").
		WithJsonTag("limit").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ListenerId").
		WithJsonTag("listener_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Marker").
		WithJsonTag("marker").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Name").
		WithJsonTag("name").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("PageReverse").
		WithJsonTag("page_reverse").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Position").
		WithJsonTag("position").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ProvisioningStatus").
		WithJsonTag("provisioning_status").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("RedirectListenerId").
		WithJsonTag("redirect_listener_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("RedirectPoolId").
		WithJsonTag("redirect_pool_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("RedirectUrl").
		WithJsonTag("redirect_url").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListL7Rules() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/l7policies/{l7policy_id}/rules").
		WithResponse(new(model.ListL7RulesResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("L7policyId").
		WithJsonTag("l7policy_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("AdminStateUp").
		WithJsonTag("admin_state_up").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("CompareType").
		WithJsonTag("compare_type").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("EnterpriseProjectId").
		WithJsonTag("enterprise_project_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Id").
		WithJsonTag("id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Invert").
		WithJsonTag("invert").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Key").
		WithJsonTag("key").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Limit").
		WithJsonTag("limit").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Marker").
		WithJsonTag("marker").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("PageReverse").
		WithJsonTag("page_reverse").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ProvisioningStatus").
		WithJsonTag("provisioning_status").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Type").
		WithJsonTag("type").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Value").
		WithJsonTag("value").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListListeners() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/listeners").
		WithResponse(new(model.ListListenersResponse)).
		WithContentType("application/json")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("AdminStateUp").
		WithJsonTag("admin_state_up").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ClientCaTlsContainerRef").
		WithJsonTag("client_ca_tls_container_ref").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ClientTimeout").
		WithJsonTag("client_timeout").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ConnectionLimit").
		WithJsonTag("connection_limit").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("DefaultPoolId").
		WithJsonTag("default_pool_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("DefaultTlsContainerRef").
		WithJsonTag("default_tls_container_ref").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Description").
		WithJsonTag("description").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("EnableMemberRetry").
		WithJsonTag("enable_member_retry").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("EnterpriseProjectId").
		WithJsonTag("enterprise_project_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Http2Enable").
		WithJsonTag("http2_enable").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Id").
		WithJsonTag("id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("KeepaliveTimeout").
		WithJsonTag("keepalive_timeout").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Limit").
		WithJsonTag("limit").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("LoadbalancerId").
		WithJsonTag("loadbalancer_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Marker").
		WithJsonTag("marker").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("MemberAddress").
		WithJsonTag("member_address").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("MemberDeviceId").
		WithJsonTag("member_device_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("MemberTimeout").
		WithJsonTag("member_timeout").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Name").
		WithJsonTag("name").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("PageReverse").
		WithJsonTag("page_reverse").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Protocol").
		WithJsonTag("protocol").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ProtocolPort").
		WithJsonTag("protocol_port").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("TlsCiphersPolicy").
		WithJsonTag("tls_ciphers_policy").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("TransparentClientIpEnable").
		WithJsonTag("transparent_client_ip_enable").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListLoadBalancers() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/loadbalancers").
		WithResponse(new(model.ListLoadBalancersResponse)).
		WithContentType("application/json")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("AdminStateUp").
		WithJsonTag("admin_state_up").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("AvailabilityZoneList").
		WithJsonTag("availability_zone_list").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("BillingInfo").
		WithJsonTag("billing_info").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("DeletionProtectionEnable").
		WithJsonTag("deletion_protection_enable").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Description").
		WithJsonTag("description").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Eips").
		WithJsonTag("eips").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("EnterpriseProjectId").
		WithJsonTag("enterprise_project_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Guaranteed").
		WithJsonTag("guaranteed").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Id").
		WithJsonTag("id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("IpVersion").
		WithJsonTag("ip_version").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Ipv6VipAddress").
		WithJsonTag("ipv6_vip_address").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Ipv6VipPortId").
		WithJsonTag("ipv6_vip_port_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Ipv6VipVirsubnetId").
		WithJsonTag("ipv6_vip_virsubnet_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("L4FlavorId").
		WithJsonTag("l4_flavor_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("L4ScaleFlavorId").
		WithJsonTag("l4_scale_flavor_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("L7FlavorId").
		WithJsonTag("l7_flavor_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("L7ScaleFlavorId").
		WithJsonTag("l7_scale_flavor_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Limit").
		WithJsonTag("limit").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Marker").
		WithJsonTag("marker").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("MemberAddress").
		WithJsonTag("member_address").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("MemberDeviceId").
		WithJsonTag("member_device_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Name").
		WithJsonTag("name").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("OperatingStatus").
		WithJsonTag("operating_status").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("PageReverse").
		WithJsonTag("page_reverse").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ProvisioningStatus").
		WithJsonTag("provisioning_status").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Publicips").
		WithJsonTag("publicips").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("VipAddress").
		WithJsonTag("vip_address").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("VipPortId").
		WithJsonTag("vip_port_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("VipSubnetCidrId").
		WithJsonTag("vip_subnet_cidr_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("VpcId").
		WithJsonTag("vpc_id").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListMembers() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/pools/{pool_id}/members").
		WithResponse(new(model.ListMembersResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("PoolId").
		WithJsonTag("pool_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Address").
		WithJsonTag("address").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("AdminStateUp").
		WithJsonTag("admin_state_up").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("EnterpriseProjectId").
		WithJsonTag("enterprise_project_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Id").
		WithJsonTag("id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Limit").
		WithJsonTag("limit").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Marker").
		WithJsonTag("marker").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Name").
		WithJsonTag("name").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("OperatingStatus").
		WithJsonTag("operating_status").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("PageReverse").
		WithJsonTag("page_reverse").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ProtocolPort").
		WithJsonTag("protocol_port").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("SubnetCidrId").
		WithJsonTag("subnet_cidr_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Weight").
		WithJsonTag("weight").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListPools() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/pools").
		WithResponse(new(model.ListPoolsResponse)).
		WithContentType("application/json")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("AdminStateUp").
		WithJsonTag("admin_state_up").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Description").
		WithJsonTag("description").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("EnterpriseProjectId").
		WithJsonTag("enterprise_project_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("HealthmonitorId").
		WithJsonTag("healthmonitor_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Id").
		WithJsonTag("id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("IpVersion").
		WithJsonTag("ip_version").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("LbAlgorithm").
		WithJsonTag("lb_algorithm").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Limit").
		WithJsonTag("limit").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("LoadbalancerId").
		WithJsonTag("loadbalancer_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Marker").
		WithJsonTag("marker").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("MemberAddress").
		WithJsonTag("member_address").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("MemberDeletionProtectionEnable").
		WithJsonTag("member_deletion_protection_enable").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("MemberDeviceId").
		WithJsonTag("member_device_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Name").
		WithJsonTag("name").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("PageReverse").
		WithJsonTag("page_reverse").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Protocol").
		WithJsonTag("protocol").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForShowCertificate() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/certificates/{certificate_id}").
		WithResponse(new(model.ShowCertificateResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("CertificateId").
		WithJsonTag("certificate_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForShowFlavor() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/flavors/{flavor_id}").
		WithResponse(new(model.ShowFlavorResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("FlavorId").
		WithJsonTag("flavor_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForShowHealthMonitor() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/healthmonitors/{healthmonitor_id}").
		WithResponse(new(model.ShowHealthMonitorResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("HealthmonitorId").
		WithJsonTag("healthmonitor_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForShowL7Policy() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/l7policies/{l7policy_id}").
		WithResponse(new(model.ShowL7PolicyResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("L7policyId").
		WithJsonTag("l7policy_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForShowL7Rule() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/l7policies/{l7policy_id}/rules/{l7rule_id}").
		WithResponse(new(model.ShowL7RuleResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("L7policyId").
		WithJsonTag("l7policy_id").
		WithLocationType(def.Path))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("L7ruleId").
		WithJsonTag("l7rule_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForShowListener() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/listeners/{listener_id}").
		WithResponse(new(model.ShowListenerResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ListenerId").
		WithJsonTag("listener_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForShowLoadBalancer() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/loadbalancers/{loadbalancer_id}").
		WithResponse(new(model.ShowLoadBalancerResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("LoadbalancerId").
		WithJsonTag("loadbalancer_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForShowLoadBalancerStatus() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/loadbalancers/{loadbalancer_id}/statuses").
		WithResponse(new(model.ShowLoadBalancerStatusResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("LoadbalancerId").
		WithJsonTag("loadbalancer_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForShowMember() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/pools/{pool_id}/members/{member_id}").
		WithResponse(new(model.ShowMemberResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("MemberId").
		WithJsonTag("member_id").
		WithLocationType(def.Path))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("PoolId").
		WithJsonTag("pool_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForShowPool() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/pools/{pool_id}").
		WithResponse(new(model.ShowPoolResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("PoolId").
		WithJsonTag("pool_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForShowQuota() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/quotas").
		WithResponse(new(model.ShowQuotaResponse)).
		WithContentType("application/json")

	// request

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForShowQuotaDefaults() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/quotas/defaults").
		WithResponse(new(model.ShowQuotaDefaultsResponse)).
		WithContentType("application/json")

	// request

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForUpdateCertificate() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPut).
		WithPath("/v3/{project_id}/elb/certificates/{certificate_id}").
		WithResponse(new(model.UpdateCertificateResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("CertificateId").
		WithJsonTag("certificate_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForUpdateHealthMonitor() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPut).
		WithPath("/v3/{project_id}/elb/healthmonitors/{healthmonitor_id}").
		WithResponse(new(model.UpdateHealthMonitorResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("HealthmonitorId").
		WithJsonTag("healthmonitor_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForUpdateL7Policy() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPut).
		WithPath("/v3/{project_id}/elb/l7policies/{l7policy_id}").
		WithResponse(new(model.UpdateL7PolicyResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("L7policyId").
		WithJsonTag("l7policy_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForUpdateL7Rule() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPut).
		WithPath("/v3/{project_id}/elb/l7policies/{l7policy_id}/rules/{l7rule_id}").
		WithResponse(new(model.UpdateL7RuleResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("L7policyId").
		WithJsonTag("l7policy_id").
		WithLocationType(def.Path))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("L7ruleId").
		WithJsonTag("l7rule_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForUpdateListener() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPut).
		WithPath("/v3/{project_id}/elb/listeners/{listener_id}").
		WithResponse(new(model.UpdateListenerResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ListenerId").
		WithJsonTag("listener_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForUpdateLoadBalancer() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPut).
		WithPath("/v3/{project_id}/elb/loadbalancers/{loadbalancer_id}").
		WithResponse(new(model.UpdateLoadBalancerResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("LoadbalancerId").
		WithJsonTag("loadbalancer_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForUpdateMember() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPut).
		WithPath("/v3/{project_id}/elb/pools/{pool_id}/members/{member_id}").
		WithResponse(new(model.UpdateMemberResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("MemberId").
		WithJsonTag("member_id").
		WithLocationType(def.Path))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("PoolId").
		WithJsonTag("pool_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForUpdatePool() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPut).
		WithPath("/v3/{project_id}/elb/pools/{pool_id}").
		WithResponse(new(model.UpdatePoolResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("PoolId").
		WithJsonTag("pool_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForCountPreoccupyIpNum() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/preoccupy-ip-num").
		WithResponse(new(model.CountPreoccupyIpNumResponse)).
		WithContentType("application/json")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("AvailabilityZoneId").
		WithJsonTag("availability_zone_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("IpTargetEnable").
		WithJsonTag("ip_target_enable").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("IpVersion").
		WithJsonTag("ip_version").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("L7FlavorId").
		WithJsonTag("l7_flavor_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("LoadbalancerId").
		WithJsonTag("loadbalancer_id").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForCreateIpGroup() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v3/{project_id}/elb/ipgroups").
		WithResponse(new(model.CreateIpGroupResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForDeleteIpGroup() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/v3/{project_id}/elb/ipgroups/{ipgroup_id}").
		WithResponse(new(model.DeleteIpGroupResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("IpgroupId").
		WithJsonTag("ipgroup_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListIpGroups() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/ipgroups").
		WithResponse(new(model.ListIpGroupsResponse)).
		WithContentType("application/json")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Description").
		WithJsonTag("description").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Id").
		WithJsonTag("id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("IpList").
		WithJsonTag("ip_list").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Limit").
		WithJsonTag("limit").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Marker").
		WithJsonTag("marker").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Name").
		WithJsonTag("name").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("PageReverse").
		WithJsonTag("page_reverse").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForShowIpGroup() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v3/{project_id}/elb/ipgroups/{ipgroup_id}").
		WithResponse(new(model.ShowIpGroupResponse)).
		WithContentType("application/json")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("IpgroupId").
		WithJsonTag("ipgroup_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForUpdateIpGroup() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPut).
		WithPath("/v3/{project_id}/elb/ipgroups/{ipgroup_id}").
		WithResponse(new(model.UpdateIpGroupResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("IpgroupId").
		WithJsonTag("ipgroup_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}
