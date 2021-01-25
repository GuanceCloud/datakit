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

// 创建密钥对请求体请求参数
type CreateKeypairAction struct {
	// SSH密钥对的名称。 - 新创建的密钥对名称不能和已有密钥对的名称相同。 - SSH密钥对名称由英文字母、数字、下划线、中划线组成，长度不能超过64个字节
	Name string `json:"name"`
	// SSH密钥对的类型
	Type *CreateKeypairActionType `json:"type,omitempty"`
	// 导入公钥的字符串信息。
	PublicKey *string `json:"public_key,omitempty"`
	// 租户级或者用户级
	Scope *CreateKeypairActionScope `json:"scope,omitempty"`
	// SSH密钥对所属的用户信息
	UserId        *string        `json:"user_id,omitempty"`
	KeyProtection *KeyProtection `json:"key_protection,omitempty"`
}

func (o CreateKeypairAction) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateKeypairAction struct{}"
	}

	return strings.Join([]string{"CreateKeypairAction", string(data)}, " ")
}

type CreateKeypairActionType struct {
	value string
}

type CreateKeypairActionTypeEnum struct {
	SSH  CreateKeypairActionType
	X509 CreateKeypairActionType
}

func GetCreateKeypairActionTypeEnum() CreateKeypairActionTypeEnum {
	return CreateKeypairActionTypeEnum{
		SSH: CreateKeypairActionType{
			value: "ssh",
		},
		X509: CreateKeypairActionType{
			value: "x509",
		},
	}
}

func (c CreateKeypairActionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateKeypairActionType) UnmarshalJSON(b []byte) error {
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

type CreateKeypairActionScope struct {
	value string
}

type CreateKeypairActionScopeEnum struct {
	DOMAIN CreateKeypairActionScope
	USER   CreateKeypairActionScope
}

func GetCreateKeypairActionScopeEnum() CreateKeypairActionScopeEnum {
	return CreateKeypairActionScopeEnum{
		DOMAIN: CreateKeypairActionScope{
			value: "domain",
		},
		USER: CreateKeypairActionScope{
			value: "user",
		},
	}
}

func (c CreateKeypairActionScope) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateKeypairActionScope) UnmarshalJSON(b []byte) error {
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
