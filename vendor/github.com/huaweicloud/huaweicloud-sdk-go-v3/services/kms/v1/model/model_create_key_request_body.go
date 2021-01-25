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

type CreateKeyRequestBody struct {
	// 非默认主密钥别名，取值范围为1到255个字符，满足正则匹配“^[a-zA-Z0-9:/_-]{1,255}$”，且不与系统服务创建的默认主密钥别名重名。
	KeyAlias *string `json:"key_alias,omitempty"`
	// 密钥描述，取值0到255字符。
	KeyDescription *string `json:"key_description,omitempty"`
	// 密钥来源，默认为“kms”，枚举如下： - kms：表示密钥材料由kms生成。 - external：表示密钥材料由外部导入。
	Origin *CreateKeyRequestBodyOrigin `json:"origin,omitempty"`
	// 企业多项目ID。 - 用户未开通企业多项目时，不需要输入该字段。 - 用户开通企业多项目时，创建资源可以输入该字段。若用户户不输入该字段，默认创建属于默认企业多项目ID（ID为“0”）的资源。 注意：若用户没有默认企业多项目ID（ID为“0”）下的创建权限，则接口报错。
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
	// 请求消息序列号，36字节序列号。 例如：919c82d4-8046-4722-9094-35c3c6524cff
	Sequence *string `json:"sequence,omitempty"`
}

func (o CreateKeyRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateKeyRequestBody struct{}"
	}

	return strings.Join([]string{"CreateKeyRequestBody", string(data)}, " ")
}

type CreateKeyRequestBodyOrigin struct {
	value string
}

type CreateKeyRequestBodyOriginEnum struct {
	KMS      CreateKeyRequestBodyOrigin
	EXTERNAL CreateKeyRequestBodyOrigin
}

func GetCreateKeyRequestBodyOriginEnum() CreateKeyRequestBodyOriginEnum {
	return CreateKeyRequestBodyOriginEnum{
		KMS: CreateKeyRequestBodyOrigin{
			value: "kms",
		},
		EXTERNAL: CreateKeyRequestBodyOrigin{
			value: "external",
		},
	}
}

func (c CreateKeyRequestBodyOrigin) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateKeyRequestBodyOrigin) UnmarshalJSON(b []byte) error {
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
