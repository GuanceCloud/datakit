/*
 * BSS
 *
 * Business Support System API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type TemplateArgs struct {
	// |参数名称：模板参数名目前仅仅支持sub_customer_name：表明企业主创建企业子的名字| |参数约束及描述：模板参数名目前仅仅支持sub_customer_name：表明企业主创建企业子的名字|
	Key string `json:"key"`
	// |参数名称：模板参数值key对应的取值| |参数约束及描述：模板参数值key对应的取值|
	Value string `json:"value"`
}

func (o TemplateArgs) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TemplateArgs struct{}"
	}

	return strings.Join([]string{"TemplateArgs", string(data)}, " ")
}
