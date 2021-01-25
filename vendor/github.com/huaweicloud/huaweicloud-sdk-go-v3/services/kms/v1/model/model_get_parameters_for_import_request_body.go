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

type GetParametersForImportRequestBody struct {
	// 密钥ID，36字节，满足正则匹配“^[0-9a-z]{8}-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{12}$”。 例如：0d0466b0-e727-4d9c-b35d-f84bb474a37f。
	KeyId *string `json:"key_id,omitempty"`
	// 密钥材料加密算法，枚举如下：  - RSAES_PKCS1_V1_5  - RSAES_OAEP_SHA_1  - RSAES_OAEP_SHA_256
	WrappingAlgorithm *GetParametersForImportRequestBodyWrappingAlgorithm `json:"wrapping_algorithm,omitempty"`
	// 请求消息序列号，36字节序列号。 例如：919c82d4-8046-4722-9094-35c3c6524cff
	Sequence *string `json:"sequence,omitempty"`
}

func (o GetParametersForImportRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "GetParametersForImportRequestBody struct{}"
	}

	return strings.Join([]string{"GetParametersForImportRequestBody", string(data)}, " ")
}

type GetParametersForImportRequestBodyWrappingAlgorithm struct {
	value string
}

type GetParametersForImportRequestBodyWrappingAlgorithmEnum struct {
	RSAES_PKCS1_V1_5   GetParametersForImportRequestBodyWrappingAlgorithm
	RSAES_OAEP_SHA_1   GetParametersForImportRequestBodyWrappingAlgorithm
	RSAES_OAEP_SHA_256 GetParametersForImportRequestBodyWrappingAlgorithm
}

func GetGetParametersForImportRequestBodyWrappingAlgorithmEnum() GetParametersForImportRequestBodyWrappingAlgorithmEnum {
	return GetParametersForImportRequestBodyWrappingAlgorithmEnum{
		RSAES_PKCS1_V1_5: GetParametersForImportRequestBodyWrappingAlgorithm{
			value: "RSAES_PKCS1_V1_5",
		},
		RSAES_OAEP_SHA_1: GetParametersForImportRequestBodyWrappingAlgorithm{
			value: "RSAES_OAEP_SHA_1",
		},
		RSAES_OAEP_SHA_256: GetParametersForImportRequestBodyWrappingAlgorithm{
			value: "RSAES_OAEP_SHA_256",
		},
	}
}

func (c GetParametersForImportRequestBodyWrappingAlgorithm) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *GetParametersForImportRequestBodyWrappingAlgorithm) UnmarshalJSON(b []byte) error {
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
