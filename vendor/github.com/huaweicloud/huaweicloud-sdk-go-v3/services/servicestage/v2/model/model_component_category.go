/*
 * ServiceStage
 *
 * ServiceStage的API,包括应用管理和仓库授权管理
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// 应用组件类型包括：Webapp、MicroService、Common。
type ComponentCategory struct {
	value string
}

type ComponentCategoryEnum struct {
	WEBAPP        ComponentCategory
	MICRO_SERVICE ComponentCategory
	COMMON        ComponentCategory
}

func GetComponentCategoryEnum() ComponentCategoryEnum {
	return ComponentCategoryEnum{
		WEBAPP: ComponentCategory{
			value: "Webapp",
		},
		MICRO_SERVICE: ComponentCategory{
			value: "MicroService",
		},
		COMMON: ComponentCategory{
			value: "Common",
		},
	}
}

func (c ComponentCategory) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ComponentCategory) UnmarshalJSON(b []byte) error {
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
