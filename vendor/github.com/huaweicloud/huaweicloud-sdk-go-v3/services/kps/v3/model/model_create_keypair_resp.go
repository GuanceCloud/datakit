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

// SSH密钥对信息详情
type CreateKeypairResp struct {
	// SSH密钥对的名称
	Name *string `json:"name,omitempty"`
	// SSH密钥对的类型
	Type *CreateKeypairRespType `json:"type,omitempty"`
	// SSH密钥对对应的publicKey信息
	PublicKey *string `json:"public_key,omitempty"`
	// SSH密钥对对应的privateKey信息 - 创建SSH密钥对时，响应中包括private_key的信息。 - 导入SSH密钥对时，响应中不包括private_key的信息。
	PrivateKey *string `json:"private_key,omitempty"`
	// SSH密钥对应指纹信息
	Fingerprint *string `json:"fingerprint,omitempty"`
	// SSH密钥对所属的用户信息
	UserId *string `json:"user_id,omitempty"`
}

func (o CreateKeypairResp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateKeypairResp struct{}"
	}

	return strings.Join([]string{"CreateKeypairResp", string(data)}, " ")
}

type CreateKeypairRespType struct {
	value string
}

type CreateKeypairRespTypeEnum struct {
	SSH  CreateKeypairRespType
	X509 CreateKeypairRespType
}

func GetCreateKeypairRespTypeEnum() CreateKeypairRespTypeEnum {
	return CreateKeypairRespTypeEnum{
		SSH: CreateKeypairRespType{
			value: "ssh",
		},
		X509: CreateKeypairRespType{
			value: "x509",
		},
	}
}

func (c CreateKeypairRespType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateKeypairRespType) UnmarshalJSON(b []byte) error {
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
