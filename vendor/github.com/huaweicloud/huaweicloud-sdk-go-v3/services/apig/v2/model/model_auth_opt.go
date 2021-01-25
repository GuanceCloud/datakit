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

// 认证方式参数
type AuthOpt struct {
	// AppCode简易认证类型，仅在auth_type为APP时生效，默认为DISABLE： - DISABLE：不开启简易认证 - HEADER：开启简易认证且AppCode位置在HEADER
	AppCodeAuthType *AuthOptAppCodeAuthType `json:"app_code_auth_type,omitempty"`
}

func (o AuthOpt) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AuthOpt struct{}"
	}

	return strings.Join([]string{"AuthOpt", string(data)}, " ")
}

type AuthOptAppCodeAuthType struct {
	value string
}

type AuthOptAppCodeAuthTypeEnum struct {
	DISABLE AuthOptAppCodeAuthType
	HEADER  AuthOptAppCodeAuthType
}

func GetAuthOptAppCodeAuthTypeEnum() AuthOptAppCodeAuthTypeEnum {
	return AuthOptAppCodeAuthTypeEnum{
		DISABLE: AuthOptAppCodeAuthType{
			value: "DISABLE",
		},
		HEADER: AuthOptAppCodeAuthType{
			value: "HEADER",
		},
	}
}

func (c AuthOptAppCodeAuthType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *AuthOptAppCodeAuthType) UnmarshalJSON(b []byte) error {
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
