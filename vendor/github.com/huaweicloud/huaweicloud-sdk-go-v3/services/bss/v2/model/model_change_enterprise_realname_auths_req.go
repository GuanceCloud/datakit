/*
 * BSS
 *
 * Business Support System API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ChangeEnterpriseRealnameAuthsReq struct {
	// |参数名称：客户ID。| |参数约束及描述：客户ID。|
	CustomerId string `json:"customer_id"`
	// |参数名称：认证方案。1：企业证件扫描| |参数的约束及描述：认证方案。1：企业证件扫描|
	IdentifyType int32 `json:"identify_type"`
	// |参数名称：企业证件类型：0：企业营业执照1：事业单位法人证书2：社会团体法人登记证书3：行政执法主体资格证4：组织机构代码证99：其他| |参数的约束及描述：企业证件类型：0：企业营业执照1：事业单位法人证书2：社会团体法人登记证书3：行政执法主体资格证4：组织机构代码证99：其他|
	CertificateType *int32 `json:"certificate_type,omitempty"`
	// |参数名称：企业证件认证时证件附件的文件URL。
	VerifiedFileUrl []string `json:"verified_file_url"`
	// |参数名称：单位名称。不能全是数字、特殊字符、空格。| |参数约束及描述：单位名称。不能全是数字、特殊字符、空格。|
	CorpName string `json:"corp_name"`
	// |参数名称：单位证件号码。| |参数约束及描述：单位证件号码。|
	VerifiedNumber string `json:"verified_number"`
	// |参数名称：实名认证填写的注册国家。国家的两位字母简码。例如：注册国家为“中国”请填写“CN”。| |参数约束及描述：实名认证填写的注册国家。国家的两位字母简码。例如：注册国家为“中国”请填写“CN”。|
	RegCountry *string `json:"reg_country,omitempty"`
	// |参数名称：实名认证企业注册地址。| |参数约束及描述：实名认证企业注册地址。|
	RegAddress *string `json:"reg_address,omitempty"`
	// |参数名称：变更类型：1：个人变企业| |参数的约束及描述：变更类型：1：个人变企业|
	ChangeType int32 `json:"change_type"`
	// |参数名称：华为分给合作伙伴的平台标识。该标识的具体值由华为分配。获取方法请参见如何获取xaccountType的取值| |参数约束及描述：华为分给合作伙伴的平台标识。该标识的具体值由华为分配。获取方法请参见如何获取xaccountType的取值|
	XaccountType     string               `json:"xaccount_type"`
	EnterprisePerson *EnterprisePersonNew `json:"enterprise_person,omitempty"`
}

func (o ChangeEnterpriseRealnameAuthsReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ChangeEnterpriseRealnameAuthsReq struct{}"
	}

	return strings.Join([]string{"ChangeEnterpriseRealnameAuthsReq", string(data)}, " ")
}
