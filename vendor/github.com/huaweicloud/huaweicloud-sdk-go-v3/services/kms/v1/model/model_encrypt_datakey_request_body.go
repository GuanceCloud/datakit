/*
 * kms
 *
 * KMS v1.0 API, open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type EncryptDatakeyRequestBody struct {
	// 密钥ID，36字节，满足正则匹配“^[0-9a-z]{8}-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{12}$”。 例如：0d0466b0-e727-4d9c-b35d-f84bb474a37f。
	KeyId *string `json:"key_id,omitempty"`
	// 一系列key-value键值对，用于记录资源上下文信息，用于保护数据的完整性，不应包含敏感信息，最大长度为8192。 当在加密时指定了该参数时，解密密文时，需要传入相同的参数，才能正确的解密。 例如：{\"Key1\":\"Value1\",\"Key2\":\"Value2\"}
	EncryptionContext *interface{} `json:"encryption_context,omitempty"`
	// DEK明文和DEK明文的SHA256（32字节），均为16进制字符串表示。 DEK明文（64字节）和DEK明文的SHA256（32字节），均为16进制字符串表示
	PlainText *string `json:"plain_text,omitempty"`
	// DEK明文字节长度，取值范围为1~1024。 DEK明文字节长度，取值为“64”。
	DatakeyPlainLength *string `json:"datakey_plain_length,omitempty"`
	// 请求消息序列号，36字节序列号。 例如：919c82d4-8046-4722-9094-35c3c6524cff
	Sequence *string `json:"sequence,omitempty"`
}

func (o EncryptDatakeyRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "EncryptDatakeyRequestBody struct{}"
	}

	return strings.Join([]string{"EncryptDatakeyRequestBody", string(data)}, " ")
}
