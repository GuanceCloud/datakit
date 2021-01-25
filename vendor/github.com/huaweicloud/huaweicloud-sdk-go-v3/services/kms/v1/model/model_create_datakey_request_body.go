/*
 * kms
 *
 * KMS v1.0 API, open API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

type CreateDatakeyRequestBody struct {
	// 密钥ID，36字节，满足正则匹配“^[0-9a-z]{8}-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{12}$”。 例如：0d0466b0-e727-4d9c-b35d-f84bb474a37f。
	KeyId *string `json:"key_id,omitempty"`
	// 一系列key-value键值对，用于记录资源上下文信息，用于保护数据的完整性，不应包含敏感信息，最大长度为8192。 当在加密时指定了该参数时，解密密文时，需要传入相同的参数，才能正确的解密。 例如：{\"Key1\":\"Value1\",\"Key2\":\"Value2\"}
	EncryptionContext *interface{} `json:"encryption_context,omitempty"`
	// 指定生成的密钥bit位长度。有效值：AES_256、AES_128。  - AES_256：表示256比特的对称密钥。  - AES_128：表示128比特的对称密钥。 说明：  datakey_length和key_spec二选一。   - 若datakey_length和key_spec都为空，默认生成256bit的密钥。   - 若datakey_length和key_spec都指定了值，仅datakey_length生效。
	KeySpec *CreateDatakeyRequestBodyKeySpec `json:"key_spec,omitempty"`
	// 密钥bit位长度。取值为8的倍数，取值范围为8~8192。 说明：  datakey_length和key_spec二选一。   - 若datakey_length和key_spec都为空，默认生成256bit的密钥。   - 若datakey_length和key_spec都指定了值，仅datakey_length生效。
	DatakeyLength *string `json:"datakey_length,omitempty"`
	// 请求消息序列号，36字节序列号。 例如：919c82d4-8046-4722-9094-35c3c6524cff
	Sequence *string `json:"sequence,omitempty"`
}

func (o CreateDatakeyRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateDatakeyRequestBody struct{}"
	}

	return strings.Join([]string{"CreateDatakeyRequestBody", string(data)}, " ")
}

type CreateDatakeyRequestBodyKeySpec struct {
	value string
}

type CreateDatakeyRequestBodyKeySpecEnum struct {
	AES_256 CreateDatakeyRequestBodyKeySpec
	AES_128 CreateDatakeyRequestBodyKeySpec
}

func GetCreateDatakeyRequestBodyKeySpecEnum() CreateDatakeyRequestBodyKeySpecEnum {
	return CreateDatakeyRequestBodyKeySpecEnum{
		AES_256: CreateDatakeyRequestBodyKeySpec{
			value: "AES_256",
		},
		AES_128: CreateDatakeyRequestBodyKeySpec{
			value: "AES_128",
		},
	}
}

func (c CreateDatakeyRequestBodyKeySpec) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateDatakeyRequestBodyKeySpec) UnmarshalJSON(b []byte) error {
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
