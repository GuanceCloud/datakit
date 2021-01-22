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

type CustomerPostalAddressV2 struct {
	// |参数名称：邮寄地址ID。| |参数约束及描述：邮寄地址ID。|
	AddressId *string `json:"address_id,omitempty"`
	// |参数名称：收件人姓名。| |参数约束及描述：收件人姓名。|
	Recipient *string `json:"recipient,omitempty"`
	// |参数名称：国家。例如：中国| |参数约束及描述：国家。例如：中国|
	Nationality *string `json:"nationality,omitempty"`
	// |参数名称：省/自治区/直辖市。例如：江苏省。| |参数约束及描述：省/自治区/直辖市。例如：江苏省。|
	Province *string `json:"province,omitempty"`
	// |参数名称：市/区。例如：南京市。| |参数约束及描述：市/区。例如：南京市。|
	City *string `json:"city,omitempty"`
	// |参数名称：区。例如：雨花区。| |参数约束及描述：区。例如：雨花区。|
	District *string `json:"district,omitempty"`
	// |参数名称：邮寄详细地址。| |参数约束及描述：邮寄详细地址。|
	Address *string `json:"address,omitempty"`
	// |参数名称：邮编。| |参数约束及描述：邮编。|
	Zipcode *string `json:"zipcode,omitempty"`
	// |参数名称：国家码，例如：中国：0086| |参数约束及描述：国家码，例如：中国：0086|
	Areacode *string `json:"areacode,omitempty"`
	// |参数名称：手机号码，不带国家码。| |参数约束及描述：手机号码，不带国家码。|
	MobilePhone *string `json:"mobile_phone,omitempty"`
	// |参数名称：是否默认地址，默认为0。1：默认地址0：非默认地址| |参数的约束及描述：是否默认地址，默认为0。1：默认地址0：非默认地址|
	IsDefault *int32 `json:"is_default,omitempty"`
}

func (o CustomerPostalAddressV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CustomerPostalAddressV2 struct{}"
	}

	return strings.Join([]string{"CustomerPostalAddressV2", string(data)}, " ")
}
