/*
 * APIG
 *
 * API网关（API Gateway）是为开发者、合作伙伴提供的高性能、高可用、高安全的API托管服务，帮助用户轻松构建、管理和发布任意规模的API。
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"
	"strings"
)

// Response Object
type ResettingAppSecretV2Response struct {
	// APP的创建者 - USER：用户自行创建 - MARKET：云市场分配
	Creator *ResettingAppSecretV2ResponseCreator `json:"creator,omitempty"`
	// 更新时间
	UpdateTime *sdktime.SdkTime `json:"update_time,omitempty"`
	// APP的key
	AppKey *string `json:"app_key,omitempty"`
	// 名称
	Name *string `json:"name,omitempty"`
	// 描述
	Remark *string `json:"remark,omitempty"`
	// 编号
	Id *string `json:"id,omitempty"`
	// 密钥
	AppSecret *string `json:"app_secret,omitempty"`
	// 注册时间
	RegisterTime *sdktime.SdkTime `json:"register_time,omitempty"`
	// 状态
	Status *int32 `json:"status,omitempty"`
	// APP的类型  默认为apig，暂不支持其他类型
	AppType        *ResettingAppSecretV2ResponseAppType `json:"app_type,omitempty"`
	HttpStatusCode int                                  `json:"-"`
}

func (o ResettingAppSecretV2Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResettingAppSecretV2Response struct{}"
	}

	return strings.Join([]string{"ResettingAppSecretV2Response", string(data)}, " ")
}

type ResettingAppSecretV2ResponseCreator struct {
	value string
}

type ResettingAppSecretV2ResponseCreatorEnum struct {
	USER   ResettingAppSecretV2ResponseCreator
	MARKET ResettingAppSecretV2ResponseCreator
}

func GetResettingAppSecretV2ResponseCreatorEnum() ResettingAppSecretV2ResponseCreatorEnum {
	return ResettingAppSecretV2ResponseCreatorEnum{
		USER: ResettingAppSecretV2ResponseCreator{
			value: "USER",
		},
		MARKET: ResettingAppSecretV2ResponseCreator{
			value: "MARKET",
		},
	}
}

func (c ResettingAppSecretV2ResponseCreator) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ResettingAppSecretV2ResponseCreator) UnmarshalJSON(b []byte) error {
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

type ResettingAppSecretV2ResponseAppType struct {
	value string
}

type ResettingAppSecretV2ResponseAppTypeEnum struct {
	APIG ResettingAppSecretV2ResponseAppType
	ROMA ResettingAppSecretV2ResponseAppType
}

func GetResettingAppSecretV2ResponseAppTypeEnum() ResettingAppSecretV2ResponseAppTypeEnum {
	return ResettingAppSecretV2ResponseAppTypeEnum{
		APIG: ResettingAppSecretV2ResponseAppType{
			value: "apig",
		},
		ROMA: ResettingAppSecretV2ResponseAppType{
			value: "roma",
		},
	}
}

func (c ResettingAppSecretV2ResponseAppType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ResettingAppSecretV2ResponseAppType) UnmarshalJSON(b []byte) error {
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
