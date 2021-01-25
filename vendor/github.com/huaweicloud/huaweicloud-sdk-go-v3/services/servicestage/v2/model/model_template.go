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

// 模板名称。
type Template struct {
	value string
}

type TemplateEnum struct {
	MAGENTO   Template
	MBAAS     Template
	WORDPRESS Template
}

func GetTemplateEnum() TemplateEnum {
	return TemplateEnum{
		MAGENTO: Template{
			value: "magento",
		},
		MBAAS: Template{
			value: "mbaas",
		},
		WORDPRESS: Template{
			value: "wordpress",
		},
	}
}

func (c Template) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *Template) UnmarshalJSON(b []byte) error {
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
