/*
 * APIG
 *
 * API网关（API Gateway）是为开发者、合作伙伴提供的高性能、高可用、高安全的API托管服务，帮助用户轻松构建、管理和发布任意规模的API。
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"
	"strings"
)

// Response Object
type UpdateSignatureKeyV2Response struct {
	// 签名密钥的密钥
	SignSecret *string `json:"sign_secret,omitempty"`
	// 更新时间
	UpdateTime *sdktime.SdkTime `json:"update_time,omitempty"`
	// 创建时间
	CreateTime *sdktime.SdkTime `json:"create_time,omitempty"`
	// 签名密钥的名称
	Name *string `json:"name,omitempty"`
	// 签名密钥的编号
	Id *string `json:"id,omitempty"`
	// 签名密钥的key
	SignKey *string `json:"sign_key,omitempty"`
	// 签名密钥类型。
	SignType       *UpdateSignatureKeyV2ResponseSignType `json:"sign_type,omitempty"`
	HttpStatusCode int                                   `json:"-"`
}

func (o UpdateSignatureKeyV2Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateSignatureKeyV2Response struct{}"
	}

	return strings.Join([]string{"UpdateSignatureKeyV2Response", string(data)}, " ")
}

type UpdateSignatureKeyV2ResponseSignType struct {
	value string
}

type UpdateSignatureKeyV2ResponseSignTypeEnum struct {
	HMAC  UpdateSignatureKeyV2ResponseSignType
	BASIC UpdateSignatureKeyV2ResponseSignType
}

func GetUpdateSignatureKeyV2ResponseSignTypeEnum() UpdateSignatureKeyV2ResponseSignTypeEnum {
	return UpdateSignatureKeyV2ResponseSignTypeEnum{
		HMAC: UpdateSignatureKeyV2ResponseSignType{
			value: "hmac",
		},
		BASIC: UpdateSignatureKeyV2ResponseSignType{
			value: "basic",
		},
	}
}

func (c UpdateSignatureKeyV2ResponseSignType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateSignatureKeyV2ResponseSignType) UnmarshalJSON(b []byte) error {
	myConverter := converter.StringConverterFactory("string")
	if myConverter != nil {
		val, err := myConverter.CovertStringToInterface(strings.Trim(string(b[:]), "\""))
		if err == nil {
			c.value = val.(string)
			return nil
		}
		return err
	} else {
		return errors.New("convert enum data to string error")
	}
}
