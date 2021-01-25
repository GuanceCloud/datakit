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

// 应用组件子类型。  Webapp的子类型有Web、Magento、Wordpress。  MicroService的子类型有Java Chassis、Go Chassis、Mesher、SpringCloud。  Common的子类型可以为空。
type ComponentSubCategory struct {
	value string
}

type ComponentSubCategoryEnum struct {
	WEB          ComponentSubCategory
	MAGENTO      ComponentSubCategory
	WORDPRESS    ComponentSubCategory
	SPRING_CLOUD ComponentSubCategory
	JAVA_CHASSIS ComponentSubCategory
	GO_CHASSIS   ComponentSubCategory
	MESHER       ComponentSubCategory
}

func GetComponentSubCategoryEnum() ComponentSubCategoryEnum {
	return ComponentSubCategoryEnum{
		WEB: ComponentSubCategory{
			value: "Web",
		},
		MAGENTO: ComponentSubCategory{
			value: "Magento",
		},
		WORDPRESS: ComponentSubCategory{
			value: "Wordpress",
		},
		SPRING_CLOUD: ComponentSubCategory{
			value: "SpringCloud",
		},
		JAVA_CHASSIS: ComponentSubCategory{
			value: "Java Chassis",
		},
		GO_CHASSIS: ComponentSubCategory{
			value: "Go Chassis",
		},
		MESHER: ComponentSubCategory{
			value: "Mesher",
		},
	}
}

func (c ComponentSubCategory) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ComponentSubCategory) UnmarshalJSON(b []byte) error {
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
