/*
 * FunctionGraph
 *
 * API v2
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
	// 函数配额限制。
	Quota *int32 `json:"quota,omitempty"`
	// 已使用的配额。
	Used *int32 `json:"used,omitempty"`
	// “资源类型”
	Type *ResourcesType `json:"type,omitempty"`
	// 资源的计数单位。fgs_func_code_size,单位为MB,其他场景无单位
	Unit *string `json:"unit,omitempty"`
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
	FGS_FUNC_SCALE_DOWN_TIMEOUT ResourcesType
	FGS_FUNC_OCCURS             ResourcesType
	FGS_FUNC_PAT_IDLE_TIME      ResourcesType
	FGS_FUNC_NUM                ResourcesType
	FGS_FUNC_CODE_SIZE          ResourcesType
	FGS_WORKFLOW_NUM            ResourcesType
}

func GetResourcesTypeEnum() ResourcesTypeEnum {
	return ResourcesTypeEnum{
		FGS_FUNC_SCALE_DOWN_TIMEOUT: ResourcesType{
			value: "fgs_func_scale_down_timeout",
		},
		FGS_FUNC_OCCURS: ResourcesType{
			value: "fgs_func_occurs",
		},
		FGS_FUNC_PAT_IDLE_TIME: ResourcesType{
			value: "fgs_func_pat_idle_time",
		},
		FGS_FUNC_NUM: ResourcesType{
			value: "fgs_func_num",
		},
		FGS_FUNC_CODE_SIZE: ResourcesType{
			value: "fgs_func_code_size",
		},
		FGS_WORKFLOW_NUM: ResourcesType{
			value: "fgs_workflow_num",
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
