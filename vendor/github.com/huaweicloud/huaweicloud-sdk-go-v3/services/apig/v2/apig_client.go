package v2

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/apig/v2/model"
)

type ApigClient struct {
	HcClient *http_client.HcHttpClient
}

func NewApigClient(hcClient *http_client.HcHttpClient) *ApigClient {
	return &ApigClient{HcClient: hcClient}
}

func ApigClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//如果创建API时，“定义API请求”使用HTTPS请求协议，那么在独立域名中需要添加SSL证书。 本章节主要介绍为特定域名绑定证书。
func (c *ApigClient) AssociateCertificateV2(request *model.AssociateCertificateV2Request) (*model.AssociateCertificateV2Response, error) {
	requestDef := GenReqDefForAssociateCertificateV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.AssociateCertificateV2Response), nil
	}
}

//用户自定义的域名，需要CNAME到API分组的子域名上才能生效，具体方法请参见[增加CNAME类型记录集](https://support.huaweicloud.com/usermanual-dns/dns_usermanual_0010.html)。 每个API分组下最多可绑定5个域名。绑定域名后，用户可通过自定义域名调用API。
func (c *ApigClient) AssociateDomainV2(request *model.AssociateDomainV2Request) (*model.AssociateDomainV2Response, error) {
	requestDef := GenReqDefForAssociateDomainV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.AssociateDomainV2Response), nil
	}
}

//签名密钥创建后，需要绑定到API才能生效。  将签名密钥绑定到API后，则API网关请求后端服务时就会使用这个签名密钥进行加密签名，后端服务可以校验这个签名来验证请求来源。  将指定的签名密钥绑定到一个或多个已发布的API上。同一个API发布到不同的环境可以绑定不同的签名密钥；一个API在发布到特定环境后只能绑定一个签名密钥。
func (c *ApigClient) AssociateSignatureKeyV2(request *model.AssociateSignatureKeyV2Request) (*model.AssociateSignatureKeyV2Response, error) {
	requestDef := GenReqDefForAssociateSignatureKeyV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.AssociateSignatureKeyV2Response), nil
	}
}

//在实际的生产中，API提供者可能有多个环境，如开发环境、测试环境、生产环境等，用户可以自由将API发布到某个环境，供调用者调用。  对于不同的环境，API的版本、请求地址甚至于包括请求消息等均有可能不同。如：某个API，v1.0的版本为稳定版本，发布到了生产环境供生产使用，同时，该API正处于迭代中，v1.1的版本是开发人员交付测试人员进行测试的版本，发布在测试环境上，而v1.2的版本目前开发团队正处于开发过程中，可以发布到开发环境进行自测等。  为此，API网关提供多环境管理功能，使租户能够最大化的模拟实际场景，低成本的接入API网关。
func (c *ApigClient) CreateEnvironmentV2(request *model.CreateEnvironmentV2Request) (*model.CreateEnvironmentV2Response, error) {
	requestDef := GenReqDefForCreateEnvironmentV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateEnvironmentV2Response), nil
	}
}

//将API发布到不同的环境后，对于不同的环境，可能会有不同的环境变量，比如，API的服务部署地址，请求的版本号等。  用户可以定义不同的环境变量，用户在定义API时，在API的定义中使用这些变量，当调用API时，API网关会将这些变量替换成真实的变量值，以达到不同环境的区分效果。  环境变量定义在API分组上，该分组下的所有API都可以使用这些变量。 > 1.环境变量的变量名称必须保持唯一，即一个分组在同一个环境上不能有两个同名的变量   2.环境变量区分大小写，即变量ABC与变量abc是两个不同的变量   3.设置了环境变量后，使用到该变量的API的调试功能将不可使用。   4.定义了环境变量后，使用到环境变量的地方应该以对称的#标识环境变量，当API发布到相应的环境后，会对环境变量的值进行替换，如：定义的API的URL为：https://#address#:8080，环境变量address在RELEASE环境上的值为：192.168.1.5，则API发布到RELEASE环境后的真实的URL为：https://192.168.1.5:8080。
func (c *ApigClient) CreateEnvironmentVariableV2(request *model.CreateEnvironmentVariableV2Request) (*model.CreateEnvironmentVariableV2Response, error) {
	requestDef := GenReqDefForCreateEnvironmentVariableV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateEnvironmentVariableV2Response), nil
	}
}

//当API上线后，系统会默认给每个API提供一个流控策略，API提供者可以根据自身API的服务能力及负载情况变更这个流控策略。 流控策略即限制API在一定长度的时间内，能够允许被访问的最大次数。
func (c *ApigClient) CreateRequestThrottlingPolicyV2(request *model.CreateRequestThrottlingPolicyV2Request) (*model.CreateRequestThrottlingPolicyV2Response, error) {
	requestDef := GenReqDefForCreateRequestThrottlingPolicyV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateRequestThrottlingPolicyV2Response), nil
	}
}

//为了保护API的安全性，建议租户为API的访问提供一套保护机制，即租户开放的API，需要对请求来源进行认证，不符合认证的请求直接拒绝访问。  其中，签名密钥就是API安全保护机制的一种。  租户创建一个签名密钥，并将签名密钥与API进行绑定，则API网关在请求这个API时，就会使用绑定的签名密钥对请求参数进行数据加密，生成签名。当租户的后端服务收到请求时，可以校验这个签名，如果签名校验不通过，则该请求不是API网关发出的请求，租户可以拒绝这个请求，从而保证API的安全性，避免API被未知来源的请求攻击。
func (c *ApigClient) CreateSignatureKeyV2(request *model.CreateSignatureKeyV2Request) (*model.CreateSignatureKeyV2Response, error) {
	requestDef := GenReqDefForCreateSignatureKeyV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateSignatureKeyV2Response), nil
	}
}

//流控策略可以限制一段时间内可以访问API的最大次数，也可以限制一段时间内单个租户和单个APP可以访问API的最大次数。  如果想要对某个特定的APP进行特殊设置，例如设置所有APP每分钟的访问次数为500次，但想设置APP1每分钟的访问次数为800次，可以通过在流控策略中设置特殊APP来实现该功能。  为流控策略添加一个特殊设置的对象，可以是APP，也可以是租户。
func (c *ApigClient) CreateSpecialThrottlingConfigurationV2(request *model.CreateSpecialThrottlingConfigurationV2Request) (*model.CreateSpecialThrottlingConfigurationV2Response, error) {
	requestDef := GenReqDefForCreateSpecialThrottlingConfigurationV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateSpecialThrottlingConfigurationV2Response), nil
	}
}

//删除指定的环境。 该操作将导致此API在指定的环境无法被访问，可能会影响相当一部分应用和用户。请确保已经告知用户，或者确认需要强制下线。
func (c *ApigClient) DeleteEnvironmentV2(request *model.DeleteEnvironmentV2Request) (*model.DeleteEnvironmentV2Response, error) {
	requestDef := GenReqDefForDeleteEnvironmentV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteEnvironmentV2Response), nil
	}
}

//删除指定的环境变量。
func (c *ApigClient) DeleteEnvironmentVariableV2(request *model.DeleteEnvironmentVariableV2Request) (*model.DeleteEnvironmentVariableV2Response, error) {
	requestDef := GenReqDefForDeleteEnvironmentVariableV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteEnvironmentVariableV2Response), nil
	}
}

//删除指定的流控策略,以及该流控策略与API的所有绑定关系。
func (c *ApigClient) DeleteRequestThrottlingPolicyV2(request *model.DeleteRequestThrottlingPolicyV2Request) (*model.DeleteRequestThrottlingPolicyV2Response, error) {
	requestDef := GenReqDefForDeleteRequestThrottlingPolicyV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteRequestThrottlingPolicyV2Response), nil
	}
}

//删除指定的签名密钥,删除签名密钥时，其配置的绑定关系会一并删除，相应的签名密钥会失效。
func (c *ApigClient) DeleteSignatureKeyV2(request *model.DeleteSignatureKeyV2Request) (*model.DeleteSignatureKeyV2Response, error) {
	requestDef := GenReqDefForDeleteSignatureKeyV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteSignatureKeyV2Response), nil
	}
}

//删除某个流控策略的某个特殊配置。
func (c *ApigClient) DeleteSpecialThrottlingConfigurationV2(request *model.DeleteSpecialThrottlingConfigurationV2Request) (*model.DeleteSpecialThrottlingConfigurationV2Response, error) {
	requestDef := GenReqDefForDeleteSpecialThrottlingConfigurationV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteSpecialThrottlingConfigurationV2Response), nil
	}
}

//如果域名证书不再需要或者已过期，则可以删除证书内容。
func (c *ApigClient) DisassociateCertificateV2(request *model.DisassociateCertificateV2Request) (*model.DisassociateCertificateV2Response, error) {
	requestDef := GenReqDefForDisassociateCertificateV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DisassociateCertificateV2Response), nil
	}
}

//如果API分组不再需要绑定某个自定义域名，则可以为此API分组解绑此域名。
func (c *ApigClient) DisassociateDomainV2(request *model.DisassociateDomainV2Request) (*model.DisassociateDomainV2Response, error) {
	requestDef := GenReqDefForDisassociateDomainV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DisassociateDomainV2Response), nil
	}
}

//解除API与签名密钥的绑定关系。
func (c *ApigClient) DisassociateSignatureKeyV2(request *model.DisassociateSignatureKeyV2Request) (*model.DisassociateSignatureKeyV2Response, error) {
	requestDef := GenReqDefForDisassociateSignatureKeyV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DisassociateSignatureKeyV2Response), nil
	}
}

//查询租户名下的API分组概况。
func (c *ApigClient) ListApiGroupsQuantitiesV2(request *model.ListApiGroupsQuantitiesV2Request) (*model.ListApiGroupsQuantitiesV2Response, error) {
	requestDef := GenReqDefForListApiGroupsQuantitiesV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListApiGroupsQuantitiesV2Response), nil
	}
}

//查询租户名下的API概况：已发布到RELEASE环境的API个数，未发布到RELEASE环境的API个数。
func (c *ApigClient) ListApiQuantitiesV2(request *model.ListApiQuantitiesV2Request) (*model.ListApiQuantitiesV2Response, error) {
	requestDef := GenReqDefForListApiQuantitiesV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListApiQuantitiesV2Response), nil
	}
}

//查询某个签名密钥上已经绑定的API列表。
func (c *ApigClient) ListApisBindedToSignatureKeyV2(request *model.ListApisBindedToSignatureKeyV2Request) (*model.ListApisBindedToSignatureKeyV2Response, error) {
	requestDef := GenReqDefForListApisBindedToSignatureKeyV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListApisBindedToSignatureKeyV2Response), nil
	}
}

//查询所有未绑定到该签名密钥上的API列表。需要API已经发布，未发布的API不予展示。
func (c *ApigClient) ListApisNotBoundWithSignatureKeyV2(request *model.ListApisNotBoundWithSignatureKeyV2Request) (*model.ListApisNotBoundWithSignatureKeyV2Response, error) {
	requestDef := GenReqDefForListApisNotBoundWithSignatureKeyV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListApisNotBoundWithSignatureKeyV2Response), nil
	}
}

//查询租户名下的APP概况：已进行API访问授权的APP个数，未进行API访问授权的APP个数。
func (c *ApigClient) ListAppQuantitiesV2(request *model.ListAppQuantitiesV2Request) (*model.ListAppQuantitiesV2Response, error) {
	requestDef := GenReqDefForListAppQuantitiesV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListAppQuantitiesV2Response), nil
	}
}

//查询分组下的所有环境变量的列表。
func (c *ApigClient) ListEnvironmentVariablesV2(request *model.ListEnvironmentVariablesV2Request) (*model.ListEnvironmentVariablesV2Response, error) {
	requestDef := GenReqDefForListEnvironmentVariablesV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListEnvironmentVariablesV2Response), nil
	}
}

//查询符合条件的环境列表。
func (c *ApigClient) ListEnvironmentsV2(request *model.ListEnvironmentsV2Request) (*model.ListEnvironmentsV2Response, error) {
	requestDef := GenReqDefForListEnvironmentsV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListEnvironmentsV2Response), nil
	}
}

//查询所有流控策略的信息。
func (c *ApigClient) ListRequestThrottlingPolicyV2(request *model.ListRequestThrottlingPolicyV2Request) (*model.ListRequestThrottlingPolicyV2Response, error) {
	requestDef := GenReqDefForListRequestThrottlingPolicyV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListRequestThrottlingPolicyV2Response), nil
	}
}

//查询某个API绑定的签名密钥列表。每个API在每个环境上应该最多只会绑定一个签名密钥。
func (c *ApigClient) ListSignatureKeysBindedToApiV2(request *model.ListSignatureKeysBindedToApiV2Request) (*model.ListSignatureKeysBindedToApiV2Response, error) {
	requestDef := GenReqDefForListSignatureKeysBindedToApiV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListSignatureKeysBindedToApiV2Response), nil
	}
}

//查询所有签名密钥的信息。
func (c *ApigClient) ListSignatureKeysV2(request *model.ListSignatureKeysV2Request) (*model.ListSignatureKeysV2Response, error) {
	requestDef := GenReqDefForListSignatureKeysV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListSignatureKeysV2Response), nil
	}
}

//查看给流控策略设置的特殊配置。
func (c *ApigClient) ListSpecialThrottlingConfigurationsV2(request *model.ListSpecialThrottlingConfigurationsV2Request) (*model.ListSpecialThrottlingConfigurationsV2Response, error) {
	requestDef := GenReqDefForListSpecialThrottlingConfigurationsV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListSpecialThrottlingConfigurationsV2Response), nil
	}
}

//查看域名下绑定的证书详情。
func (c *ApigClient) ShowDetailsOfDomainNameCertificateV2(request *model.ShowDetailsOfDomainNameCertificateV2Request) (*model.ShowDetailsOfDomainNameCertificateV2Response, error) {
	requestDef := GenReqDefForShowDetailsOfDomainNameCertificateV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowDetailsOfDomainNameCertificateV2Response), nil
	}
}

//查看指定的环境变量的详情。
func (c *ApigClient) ShowDetailsOfEnvironmentVariableV2(request *model.ShowDetailsOfEnvironmentVariableV2Request) (*model.ShowDetailsOfEnvironmentVariableV2Response, error) {
	requestDef := GenReqDefForShowDetailsOfEnvironmentVariableV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowDetailsOfEnvironmentVariableV2Response), nil
	}
}

//查看指定流控策略的详细信息。
func (c *ApigClient) ShowDetailsOfRequestThrottlingPolicyV2(request *model.ShowDetailsOfRequestThrottlingPolicyV2Request) (*model.ShowDetailsOfRequestThrottlingPolicyV2Response, error) {
	requestDef := GenReqDefForShowDetailsOfRequestThrottlingPolicyV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowDetailsOfRequestThrottlingPolicyV2Response), nil
	}
}

//修改指定环境的信息。其中可修改的属性为：name、remark，其它属性不可修改。
func (c *ApigClient) UpdateEnvironmentV2(request *model.UpdateEnvironmentV2Request) (*model.UpdateEnvironmentV2Response, error) {
	requestDef := GenReqDefForUpdateEnvironmentV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateEnvironmentV2Response), nil
	}
}

//修改指定流控策略的详细信息。
func (c *ApigClient) UpdateRequestThrottlingPolicyV2(request *model.UpdateRequestThrottlingPolicyV2Request) (*model.UpdateRequestThrottlingPolicyV2Response, error) {
	requestDef := GenReqDefForUpdateRequestThrottlingPolicyV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateRequestThrottlingPolicyV2Response), nil
	}
}

//修改指定签名密钥的详细信息。
func (c *ApigClient) UpdateSignatureKeyV2(request *model.UpdateSignatureKeyV2Request) (*model.UpdateSignatureKeyV2Response, error) {
	requestDef := GenReqDefForUpdateSignatureKeyV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateSignatureKeyV2Response), nil
	}
}

//修改某个流控策略下的某个特殊设置。
func (c *ApigClient) UpdateSpecialThrottlingConfigurationV2(request *model.UpdateSpecialThrottlingConfigurationV2Request) (*model.UpdateSpecialThrottlingConfigurationV2Response, error) {
	requestDef := GenReqDefForUpdateSpecialThrottlingConfigurationV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateSpecialThrottlingConfigurationV2Response), nil
	}
}

//将流控策略应用于API，则所有对该API的访问将会受到该流控策略的限制。  当一定时间内的访问次数超过流控策略设置的API最大访问次数限制后，后续的访问将会被拒绝，从而能够较好的保护后端API免受异常流量的冲击，保障服务的稳定运行。  为指定的API绑定流控策略，绑定时，需要指定在哪个环境上生效。同一个API发布到不同的环境可以绑定不同的流控策略；一个API在发布到特定环境后只能绑定一个默认的流控策略。
func (c *ApigClient) AssociateRequestThrottlingPolicyV2(request *model.AssociateRequestThrottlingPolicyV2Request) (*model.AssociateRequestThrottlingPolicyV2Response, error) {
	requestDef := GenReqDefForAssociateRequestThrottlingPolicyV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.AssociateRequestThrottlingPolicyV2Response), nil
	}
}

//批量解除API与流控策略的绑定关系
func (c *ApigClient) BatchDisassociateThrottlingPolicyV2(request *model.BatchDisassociateThrottlingPolicyV2Request) (*model.BatchDisassociateThrottlingPolicyV2Response, error) {
	requestDef := GenReqDefForBatchDisassociateThrottlingPolicyV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.BatchDisassociateThrottlingPolicyV2Response), nil
	}
}

//API分组是API的管理单元，一个API分组等同于一个服务入口，创建API分组时，返回一个子域名作为访问入口。建议一个API分组下的API具有一定的相关性。
func (c *ApigClient) CreateApiGroupV2(request *model.CreateApiGroupV2Request) (*model.CreateApiGroupV2Response, error) {
	requestDef := GenReqDefForCreateApiGroupV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateApiGroupV2Response), nil
	}
}

//添加一个API，API即一个服务接口，具体的服务能力。  API分为两部分，第一部分为面向API使用者的API接口，定义了使用者如何调用这个API。第二部分面向API提供者，由API提供者定义这个API的真实的后端情况，定义了API网关如何去访问真实的后端服务。API的真实后端服务目前支持三种类型：传统的HTTP/HTTPS形式的web后端、函数工作流、MOCK。
func (c *ApigClient) CreateApiV2(request *model.CreateApiV2Request) (*model.CreateApiV2Response, error) {
	requestDef := GenReqDefForCreateApiV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateApiV2Response), nil
	}
}

//对API进行发布或下线。  发布操作是将一个指定的API发布到一个指定的环境，API只有发布后，才能够被调用，且只能在该环境上才能被调用。未发布的API无法被调用。  下线操作是将API从某个已发布的环境上下线，下线后，API将无法再被调用。
func (c *ApigClient) CreateOrDeletePublishRecordForApiV2(request *model.CreateOrDeletePublishRecordForApiV2Request) (*model.CreateOrDeletePublishRecordForApiV2Response, error) {
	requestDef := GenReqDefForCreateOrDeletePublishRecordForApiV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateOrDeletePublishRecordForApiV2Response), nil
	}
}

//删除指定的API分组。  删除时，会一并删除直接或间接关联到该分组下的所有资源，包括API、独立域名、SSL证书、上架信息、分组下所有API的授权信息、编排信息、白名单配置、认证增强信息等等。并会将外部域名与子域名的绑定关系进行解除（取决于域名cname方式）。
func (c *ApigClient) DeleteApiGroupV2(request *model.DeleteApiGroupV2Request) (*model.DeleteApiGroupV2Response, error) {
	requestDef := GenReqDefForDeleteApiGroupV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteApiGroupV2Response), nil
	}
}

//删除指定的API。  删除API时，会删除该API所有相关的资源信息或绑定关系，如API的发布记录，绑定的后端服务，对APP的授权信息等。
func (c *ApigClient) DeleteApiV2(request *model.DeleteApiV2Request) (*model.DeleteApiV2Response, error) {
	requestDef := GenReqDefForDeleteApiV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteApiV2Response), nil
	}
}

//解除API与流控策略的绑定关系。
func (c *ApigClient) DisassociateRequestThrottlingPolicyV2(request *model.DisassociateRequestThrottlingPolicyV2Request) (*model.DisassociateRequestThrottlingPolicyV2Response, error) {
	requestDef := GenReqDefForDisassociateRequestThrottlingPolicyV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DisassociateRequestThrottlingPolicyV2Response), nil
	}
}

//查询API分组列表。  如果是租户操作，则查询该租户下所有的分组；如果是管理员操作，则查询的是所有租户的分组。
func (c *ApigClient) ListApiGroupsV2(request *model.ListApiGroupsV2Request) (*model.ListApiGroupsV2Response, error) {
	requestDef := GenReqDefForListApiGroupsV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListApiGroupsV2Response), nil
	}
}

//查询某个流控策略上已经绑定的API列表。
func (c *ApigClient) ListApisBindedToRequestThrottlingPolicyV2(request *model.ListApisBindedToRequestThrottlingPolicyV2Request) (*model.ListApisBindedToRequestThrottlingPolicyV2Response, error) {
	requestDef := GenReqDefForListApisBindedToRequestThrottlingPolicyV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListApisBindedToRequestThrottlingPolicyV2Response), nil
	}
}

//查询所有未绑定到该流控策略上的自有API列表。需要API已经发布，未发布的API不予展示。
func (c *ApigClient) ListApisUnbindedToRequestThrottlingPolicyV2(request *model.ListApisUnbindedToRequestThrottlingPolicyV2Request) (*model.ListApisUnbindedToRequestThrottlingPolicyV2Response, error) {
	requestDef := GenReqDefForListApisUnbindedToRequestThrottlingPolicyV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListApisUnbindedToRequestThrottlingPolicyV2Response), nil
	}
}

//查看API列表，返回API详细信息、发布信息等，但不能查看到后端服务信息。
func (c *ApigClient) ListApisV2(request *model.ListApisV2Request) (*model.ListApisV2Response, error) {
	requestDef := GenReqDefForListApisV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListApisV2Response), nil
	}
}

//查询某个API绑定的流控策略列表。每个环境上应该最多只有一个流控策略。
func (c *ApigClient) ListRequestThrottlingPoliciesBindedToApiV2(request *model.ListRequestThrottlingPoliciesBindedToApiV2Request) (*model.ListRequestThrottlingPoliciesBindedToApiV2Response, error) {
	requestDef := GenReqDefForListRequestThrottlingPoliciesBindedToApiV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListRequestThrottlingPoliciesBindedToApiV2Response), nil
	}
}

//查询指定分组的详细信息。
func (c *ApigClient) ShowDetailsOfApiGroupV2(request *model.ShowDetailsOfApiGroupV2Request) (*model.ShowDetailsOfApiGroupV2Response, error) {
	requestDef := GenReqDefForShowDetailsOfApiGroupV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowDetailsOfApiGroupV2Response), nil
	}
}

//查看指定的API的详细信息。
func (c *ApigClient) ShowDetailsOfApiV2(request *model.ShowDetailsOfApiV2Request) (*model.ShowDetailsOfApiV2Response, error) {
	requestDef := GenReqDefForShowDetailsOfApiV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowDetailsOfApiV2Response), nil
	}
}

//修改API分组属性。其中name和remark可修改，其他属性不可修改。
func (c *ApigClient) UpdateApiGroupV2(request *model.UpdateApiGroupV2Request) (*model.UpdateApiGroupV2Response, error) {
	requestDef := GenReqDefForUpdateApiGroupV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateApiGroupV2Response), nil
	}
}

//修改指定API的信息，包括后端服务信息。
func (c *ApigClient) UpdateApiV2(request *model.UpdateApiV2Request) (*model.UpdateApiV2Response, error) {
	requestDef := GenReqDefForUpdateApiV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateApiV2Response), nil
	}
}

//解除API对APP的授权关系。解除授权后，APP将不再能够调用该API。
func (c *ApigClient) CancelingAuthorizationV2(request *model.CancelingAuthorizationV2Request) (*model.CancelingAuthorizationV2Response, error) {
	requestDef := GenReqDefForCancelingAuthorizationV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CancelingAuthorizationV2Response), nil
	}
}

//校验app是否存在，非APP所有者可以调用该接口校验APP是否真实存在。这个接口只展示app的基本信息id 、name、 remark，其他信息不显示。
func (c *ApigClient) CheckAppV2(request *model.CheckAppV2Request) (*model.CheckAppV2Response, error) {
	requestDef := GenReqDefForCheckAppV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CheckAppV2Response), nil
	}
}

//APP即应用，是一个可以访问API的身份标识。将API授权给APP后，APP即可调用API。 创建一个APP。
func (c *ApigClient) CreateAnAppV2(request *model.CreateAnAppV2Request) (*model.CreateAnAppV2Response, error) {
	requestDef := GenReqDefForCreateAnAppV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateAnAppV2Response), nil
	}
}

//APP创建成功后，还不能访问API，如果想要访问某个环境上的API，需要将该API在该环境上授权给APP。授权成功后，APP即可访问该环境上的这个API。
func (c *ApigClient) CreateAuthorizingAppsV2(request *model.CreateAuthorizingAppsV2Request) (*model.CreateAuthorizingAppsV2Response, error) {
	requestDef := GenReqDefForCreateAuthorizingAppsV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateAuthorizingAppsV2Response), nil
	}
}

//删除指定的APP。 APP删除后，将无法再调用任何API；其中，云市场自动创建的APP无法被删除。
func (c *ApigClient) DeleteAppV2(request *model.DeleteAppV2Request) (*model.DeleteAppV2Response, error) {
	requestDef := GenReqDefForDeleteAppV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteAppV2Response), nil
	}
}

//查询APP已经绑定的API列表。
func (c *ApigClient) ListApisBindedToAppV2(request *model.ListApisBindedToAppV2Request) (*model.ListApisBindedToAppV2Response, error) {
	requestDef := GenReqDefForListApisBindedToAppV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListApisBindedToAppV2Response), nil
	}
}

//查询指定环境上某个APP未绑定的API列表，包括自有API和从云市场购买的API。
func (c *ApigClient) ListApisUnbindedToAppV2(request *model.ListApisUnbindedToAppV2Request) (*model.ListApisUnbindedToAppV2Response, error) {
	requestDef := GenReqDefForListApisUnbindedToAppV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListApisUnbindedToAppV2Response), nil
	}
}

//查询API绑定的APP列表。
func (c *ApigClient) ListAppsBindedToApiV2(request *model.ListAppsBindedToApiV2Request) (*model.ListAppsBindedToApiV2Response, error) {
	requestDef := GenReqDefForListAppsBindedToApiV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListAppsBindedToApiV2Response), nil
	}
}

//查询APP列表。
func (c *ApigClient) ListAppsV2(request *model.ListAppsV2Request) (*model.ListAppsV2Response, error) {
	requestDef := GenReqDefForListAppsV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListAppsV2Response), nil
	}
}

//重置指定APP的密钥。
func (c *ApigClient) ResettingAppSecretV2(request *model.ResettingAppSecretV2Request) (*model.ResettingAppSecretV2Response, error) {
	requestDef := GenReqDefForResettingAppSecretV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ResettingAppSecretV2Response), nil
	}
}

//查看指定APP的详细信息。
func (c *ApigClient) ShowDetailsOfAppV2(request *model.ShowDetailsOfAppV2Request) (*model.ShowDetailsOfAppV2Response, error) {
	requestDef := GenReqDefForShowDetailsOfAppV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowDetailsOfAppV2Response), nil
	}
}

//修改指定APP的信息。其中可修改的属性为：name、remark，当支持用户自定义key和secret的开关开启时，app_key和app_secret也支持修改，其它属性不可修改。
func (c *ApigClient) UpdateAppV2(request *model.UpdateAppV2Request) (*model.UpdateAppV2Response, error) {
	requestDef := GenReqDefForUpdateAppV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateAppV2Response), nil
	}
}
