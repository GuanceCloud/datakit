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
type UpdateAppV2Response struct {
	// APP的创建者 - USER：用户自行创建 - MARKET：云市场分配
	Creator *UpdateAppV2ResponseCreator `json:"creator,omitempty"`
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
	AppType        *UpdateAppV2ResponseAppType `json:"app_type,omitempty"`
	HttpStatusCode int                         `json:"-"`
}

func (o UpdateAppV2Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateAppV2Response struct{}"
	}

	return strings.Join([]string{"UpdateAppV2Response", string(data)}, " ")
}

type UpdateAppV2ResponseCreator struct {
	value string
}

type UpdateAppV2ResponseCreatorEnum struct {
	USER   UpdateAppV2ResponseCreator
	MARKET UpdateAppV2ResponseCreator
}

func GetUpdateAppV2ResponseCreatorEnum() UpdateAppV2ResponseCreatorEnum {
	return UpdateAppV2ResponseCreatorEnum{
		USER: UpdateAppV2ResponseCreator{
			value: "USER",
		},
		MARKET: UpdateAppV2ResponseCreator{
			value: "MARKET",
		},
	}
}

func (c UpdateAppV2ResponseCreator) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateAppV2ResponseCreator) UnmarshalJSON(b []byte) error {
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

type UpdateAppV2ResponseAppType struct {
	value string
}

type UpdateAppV2ResponseAppTypeEnum struct {
	APIG UpdateAppV2ResponseAppType
	ROMA UpdateAppV2ResponseAppType
}

func GetUpdateAppV2ResponseAppTypeEnum() UpdateAppV2ResponseAppTypeEnum {
	return UpdateAppV2ResponseAppTypeEnum{
		APIG: UpdateAppV2ResponseAppType{
			value: "apig",
		},
		ROMA: UpdateAppV2ResponseAppType{
			value: "roma",
		},
	}
}

func (c UpdateAppV2ResponseAppType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateAppV2ResponseAppType) UnmarshalJSON(b []byte) error {
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
