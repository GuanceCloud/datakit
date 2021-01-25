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

type MeasureUnitRest struct {
	// |参数名称：度量单位ID| |参数的约束及描述：度量单位ID|
	MeasureId *int32 `json:"measure_id,omitempty"`
	// |参数名称：度量单位名称（默认语言或者要查询语言名称）| |参数约束及描述：度量单位名称（默认语言或者要查询语言名称）|
	MeasureName *string `json:"measure_name,omitempty"`
	// |参数名称：英文缩写| |参数约束及描述：英文缩写|
	Abbreviation *string `json:"abbreviation,omitempty"`
	// |参数名称：度量类型| |参数的约束及描述：度量类型|
	MeasureType *int32 `json:"measure_type,omitempty"`
}

func (o MeasureUnitRest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "MeasureUnitRest struct{}"
	}

	return strings.Join([]string{"MeasureUnitRest", string(data)}, " ")
}
