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

type DecryptDatakeyRequestBody struct {
	// 密钥ID，36字节，满足正则匹配“^[0-9a-z]{8}-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{12}$”。 例如：0d0466b0-e727-4d9c-b35d-f84bb474a37f。
	KeyId *string `json:"key_id,omitempty"`
	// 一系列key-value键值对，用于记录资源上下文信息，用于保护数据的完整性，不应包含敏感信息，最大长度为8192。 当在加密时指定了该参数时，解密密文时，需要传入相同的参数，才能正确的解密。 例如：{\"Key1\":\"Value1\",\"Key2\":\"Value2\"}
	EncryptionContext *interface{} `json:"encryption_context,omitempty"`
	// DEK密文及元数据的16进制字符串。取值为加密数据密钥结果中的cipher_text的值。
	CipherText *string `json:"cipher_text,omitempty"`
	// 密钥字节长度，取值范围为1~1024。 密钥字节长度，取值为“64”。
	DatakeyCipherLength *string `json:"datakey_cipher_length,omitempty"`
	// 请求消息序列号，36字节序列号。 例如：919c82d4-8046-4722-9094-35c3c6524cff
	Sequence *string `json:"sequence,omitempty"`
}

func (o DecryptDatakeyRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DecryptDatakeyRequestBody struct{}"
	}

	return strings.Join([]string{"DecryptDatakeyRequestBody", string(data)}, " ")
}
