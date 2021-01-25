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

type ServiceType struct {
	// |参数名称：云服务类型名称| |参数约束及描述：云服务类型名称|
	ServiceTypeName *string `json:"service_type_name,omitempty"`
	// |参数名称：云服务类型编码| |参数约束及描述：云服务类型编码|
	ServiceTypeCode *string `json:"service_type_code,omitempty"`
	// |参数名称：云服务缩写| |参数约束及描述：云服务缩写|
	Abbreviation *string `json:"abbreviation,omitempty"`
}

func (o ServiceType) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ServiceType struct{}"
	}

	return strings.Join([]string{"ServiceType", string(data)}, " ")
}
