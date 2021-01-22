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

type CustomerInformation struct {
	// |参数名称：实名认证名称。虚拟账号下，该字段无效。| |参数约束及描述：实名认证名称。虚拟账号下，该字段无效。|
	Customer *string `json:"customer,omitempty"`
	// |参数名称：客户登录名称（如果客户创建了子用户，此处返回主账号登录名称）。| |参数约束及描述：客户登录名称（如果客户创建了子用户，此处返回主账号登录名称）。|
	AccountName string `json:"account_name"`
	// |参数名称：客户ID。| |参数约束及描述：客户ID。|
	CustomerId string `json:"customer_id"`
	// |参数名称：客户和伙伴关联时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”，其中，HH范围是0～23，mm和ss范围是0～59。| |参数约束及描述：客户和伙伴关联时间。UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”，其中，HH范围是0～23，mm和ss范围是0～59。|
	AssociatedOn *string `json:"associated_on,omitempty"`
	// |参数名称：合作模式。1：推荐2：垫付3：转售| |参数约束及描述：合作模式。1：推荐2：垫付3：转售|
	AssociationType *string `json:"association_type,omitempty"`
	// |参数名称：标签，支持模糊查找。虚拟账号下，该字段无效。| |参数约束及描述：标签，支持模糊查找。虚拟账号下，该字段无效。|
	Label *string `json:"label,omitempty"`
	// |参数名称：客户电话号码。虚拟账号下，该字段无效。| |参数约束及描述：客户电话号码。虚拟账号下，该字段无效。|
	Telephone *string `json:"telephone,omitempty"`
	// |参数名称：实名认证状态，虚拟账号下，该字段无效。：null：实名认证开关关闭；-1：未实名认证；0：实名认证审核中；1：实名认证不通过；2：已实名认证；3：实名认证失败。| |参数约束及描述：实名认证状态，虚拟账号下，该字段无效。：null：实名认证开关关闭；-1：未实名认证；0：实名认证审核中；1：实名认证不通过；2：已实名认证；3：实名认证失败。|
	VerifiedStatus *string `json:"verified_status,omitempty"`
	// |参数名称：国家码，电话号码的国家码前缀。虚拟账号下，该字段无效。例如：中国 0086。| |参数约束及描述：国家码，电话号码的国家码前缀。虚拟账号下，该字段无效。例如：中国 0086。|
	CountryCode *string `json:"country_code,omitempty"`
	// |参数名称：客户类型，虚拟账号下，该字段无效。：-1：无类型0：个人1：企业客户刚注册的时候，没有具体的客户类型，为“-1：无类型”，客户可以在账号中心通过设置客户类型或者在实名认证的时候，选择对应的企业/个人实名认证来决定自己的类型。| |参数的约束及描述：客户类型，虚拟账号下，该字段无效。：-1：无类型0：个人1：企业客户刚注册的时候，没有具体的客户类型，为“-1：无类型”，客户可以在账号中心通过设置客户类型或者在实名认证的时候，选择对应的企业/个人实名认证来决定自己的类型。|
	CustomerType *int32 `json:"customer_type,omitempty"`
	// |参数名称：是否伙伴冻结，注意，只有转售子客户才能被伙伴冻结：0：否1：是| |参数的约束及描述：是否伙伴冻结，注意，只有转售子客户才能被伙伴冻结：0：否1：是|
	IsFrozen *int32 `json:"is_frozen,omitempty"`
	// |参数名称：客户经理名称列表，目前只支持1个| |参数约束以及描述：客户经理名称列表，目前只支持1个|
	AccountManagers *[]AccountManager `json:"account_managers,omitempty"`
}

func (o CustomerInformation) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CustomerInformation struct{}"
	}

	return strings.Join([]string{"CustomerInformation", string(data)}, " ")
}
