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

// 密钥对信息
type Keypair struct {
	// SSH密钥对的名称
	Name *string `json:"name,omitempty"`
	// SSH密钥对的类型，值为“ssh”或“x509”
	Type *KeypairType `json:"type,omitempty"`
	// 租户级或者用户级
	Scope *KeypairScope `json:"scope,omitempty"`
	// SSH密钥对对应的publicKey信息
	PublicKey *string `json:"public_key,omitempty"`
	// SSH密钥对应指纹信息
	Fingerprint *string `json:"fingerprint,omitempty"`
	// 是否托管密钥
	IsKeyProtection *bool `json:"is_key_protection,omitempty"`
	// 冻结状态 - 0：正常状态 - 1：普通冻结 - 2：公安冻结 - 3：普通冻结及公安冻结 - 4：违规冻结 - 5：普通冻结及违规冻结 - 6：公安冻结及违规冻结 - 7：普通冻结、公安冻结及违规冻结 - 8：未实名认证冻结 - 9：普通冻结及未实名认证冻结 - 10：公安冻结及未实名认证冻结
	FrozenState *string `json:"frozen_state,omitempty"`
}

func (o Keypair) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Keypair struct{}"
	}

	return strings.Join([]string{"Keypair", string(data)}, " ")
}

type KeypairType struct {
	value string
}

type KeypairTypeEnum struct {
	SSH  KeypairType
	X509 KeypairType
}

func GetKeypairTypeEnum() KeypairTypeEnum {
	return KeypairTypeEnum{
		SSH: KeypairType{
			value: "ssh",
		},
		X509: KeypairType{
			value: "x509",
		},
	}
}

func (c KeypairType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *KeypairType) UnmarshalJSON(b []byte) error {
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

type KeypairScope struct {
	value string
}

type KeypairScopeEnum struct {
	DOMAIN KeypairScope
	USER   KeypairScope
}

func GetKeypairScopeEnum() KeypairScopeEnum {
	return KeypairScopeEnum{
		DOMAIN: KeypairScope{
			value: "domain",
		},
		USER: KeypairScope{
			value: "user",
		},
	}
}

func (c KeypairScope) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *KeypairScope) UnmarshalJSON(b []byte) error {
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
