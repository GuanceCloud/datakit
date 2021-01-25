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

type CreatePartnerCouponsReq struct {
	// |参数名称：优惠券额度ID优惠券的类型跟随额度中的类型。| |参数约束及描述：优惠券额度ID优惠券的类型跟随额度中的类型。|
	QuotaId string `json:"quota_id"`
	// |参数名称：客户ID列表| |参数约束以及描述：客户ID列表|
	CustomerIds []string `json:"customer_ids"`
	// |参数名称：优惠券的面值：小数点后2位。浮点数精度为：小数点后两位| |参数的约束及描述：优惠券的面值：小数点后2位|
	FaceValue float32 `json:"face_value"`
	// |参数名称：优惠券的生效时间,UTC格式：yyyy-MM-ddTHH:mm:ssZ| |参数约束及描述：优惠券的生效时间,UTC格式：yyyy-MM-ddTHH:mm:ssZ|
	ValidTime *string `json:"valid_time,omitempty"`
	// |参数名称：优惠券的失效时间,UTC格式：yyyy-MM-ddTHH:mm:ssZ| |参数约束及描述：优惠券的失效时间,UTC格式：yyyy-MM-ddTHH:mm:ssZ|
	ExpireTime *string `json:"expire_time,omitempty"`
	// |参数名称：云服务限制| |参数约束以及描述：云服务限制|
	CloudServiceTypes *[]string `json:"cloud_service_types,omitempty"`
	// |参数名称：产品限制| |参数约束以及描述：产品限制|
	ProductIds *[]string `json:"product_ids,omitempty"`
	// |参数名称：发券时的备注信息| |参数约束及描述：发券时的备注信息|
	Memo *string `json:"memo,omitempty"`
	// |参数名称：二级经销商ID| |参数约束及描述：如果一级经销商要给二级经销商的子客户设置折扣，需要携带这个字段|
	IndirectPartnerId *string `json:"indirect_partner_id,omitempty"`
}

func (o CreatePartnerCouponsReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreatePartnerCouponsReq struct{}"
	}

	return strings.Join([]string{"CreatePartnerCouponsReq", string(data)}, " ")
}
