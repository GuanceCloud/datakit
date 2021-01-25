/*
 * kps
 *
 * kps v3 版本API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// 可选字段，鉴权认证类型。替换时需要该参数，重置时不需要该参数。
type Auth struct {
	// 取值为枚举类型。
	Type *AuthType `json:"type,omitempty"`
	// - type为枚举值password时，key表示密码； - type为枚举值keypair时，key表示私钥；
	Key *string `json:"key,omitempty"`
}

func (o Auth) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Auth struct{}"
	}

	return strings.Join([]string{"Auth", string(data)}, " ")
}

type AuthType struct {
	value string
}

type AuthTypeEnum struct {
	PASSWORD AuthType
	KEYPAIR  AuthType
}

func GetAuthTypeEnum() AuthTypeEnum {
	return AuthTypeEnum{
		PASSWORD: AuthType{
			value: "password",
		},
		KEYPAIR: AuthType{
			value: "keypair",
		},
	}
}

func (c AuthType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *AuthType) UnmarshalJSON(b []byte) error {
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
