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

type Resources struct {
	// 配额类型。枚举值说明:  - CMK，用户主密钥  - grant_per_CMK，单个用户主密钥可创建授权数
	Type *ResourcesType `json:"type,omitempty"`
	// 已使用配额数。
	Used *int32 `json:"used,omitempty"`
	// 配额总数。
	Quota *int32 `json:"quota,omitempty"`
}

func (o Resources) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Resources struct{}"
	}

	return strings.Join([]string{"Resources", string(data)}, " ")
}

type ResourcesType struct {
	value string
}

type ResourcesTypeEnum struct {
	CMK           ResourcesType
	GRANT_PER_CMK ResourcesType
}

func GetResourcesTypeEnum() ResourcesTypeEnum {
	return ResourcesTypeEnum{
		CMK: ResourcesType{
			value: "CMK",
		},
		GRANT_PER_CMK: ResourcesType{
			value: "grant_per_CMK",
		},
	}
}

func (c ResourcesType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ResourcesType) UnmarshalJSON(b []byte) error {
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
