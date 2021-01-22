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
	"strings"
)

type ApiPolicyReqBase struct {
	// 关联的策略组合模式： - ALL：满足全部条件 - ANY：满足任一条件
	EffectMode ApiPolicyReqBaseEffectMode `json:"effect_mode"`
	// 策略后端名称。字符串由中文、英文字母、数字、下划线组成，且只能以中文或英文开头。
	Name string `json:"name"`
	// 后端参数列表
	BackendParams *[]BackendParamBase `json:"backend_params,omitempty"`
	// 策略条件列表
	Conditions []ApiConditionBase `json:"conditions"`
	// 后端自定义认证对象的ID
	AuthorizerId *string `json:"authorizer_id,omitempty"`
}

func (o ApiPolicyReqBase) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ApiPolicyReqBase struct{}"
	}

	return strings.Join([]string{"ApiPolicyReqBase", string(data)}, " ")
}

type ApiPolicyReqBaseEffectMode struct {
	value string
}

type ApiPolicyReqBaseEffectModeEnum struct {
	ALL ApiPolicyReqBaseEffectMode
	ANY ApiPolicyReqBaseEffectMode
}

func GetApiPolicyReqBaseEffectModeEnum() ApiPolicyReqBaseEffectModeEnum {
	return ApiPolicyReqBaseEffectModeEnum{
		ALL: ApiPolicyReqBaseEffectMode{
			value: "ALL",
		},
		ANY: ApiPolicyReqBaseEffectMode{
			value: "ANY",
		},
	}
}

func (c ApiPolicyReqBaseEffectMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ApiPolicyReqBaseEffectMode) UnmarshalJSON(b []byte) error {
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
