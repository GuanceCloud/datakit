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
	"strings"
)

type SignatureReq struct {
	// 签名密钥的密钥。支持英文，数字，下划线，中划线，!，@，#，$，%，且只能以英文字母开头，16 ~ 64字符。未填写时后台自动生成。
	SignSecret *string `json:"sign_secret,omitempty"`
	// 签名密钥的名称。支持汉字，英文，数字，下划线，且只能以英文和汉字开头，3 ~ 64字符。 > 中文字符必须为UTF-8或者unicode编码。
	Name string `json:"name"`
	// 签名密钥的key。支持英文，数字，下划线，中划线，且只能以英文字母开头，8 ~ 32字符。未填写时后台自动生成。
	SignKey *string `json:"sign_key,omitempty"`
	// 签名密钥类型。
	SignType *SignatureReqSignType `json:"sign_type,omitempty"`
}

func (o SignatureReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SignatureReq struct{}"
	}

	return strings.Join([]string{"SignatureReq", string(data)}, " ")
}

type SignatureReqSignType struct {
	value string
}

type SignatureReqSignTypeEnum struct {
	HMAC  SignatureReqSignType
	BASIC SignatureReqSignType
}

func GetSignatureReqSignTypeEnum() SignatureReqSignTypeEnum {
	return SignatureReqSignTypeEnum{
		HMAC: SignatureReqSignType{
			value: "hmac",
		},
		BASIC: SignatureReqSignType{
			value: "basic",
		},
	}
}

func (c SignatureReqSignType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *SignatureReqSignType) UnmarshalJSON(b []byte) error {
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
