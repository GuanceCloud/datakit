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

type Conversion struct {
	// |参数名称：度量单位| |参数的约束及描述：度量单位|
	MeasureId *int32 `json:"measure_id,omitempty"`
	// |参数名称：转换的度量单位| |参数的约束及描述：转换的度量单位|
	RefMeasureId *int32 `json:"ref_measure_id,omitempty"`
	// |参数名称：转换比率| |参数的约束及描述：转换比率|
	ConversionRatio *int64 `json:"conversion_ratio,omitempty"`
	// |参数名称：度量类型| |参数的约束及描述：度量类型|
	MeasureType *int32 `json:"measure_type,omitempty"`
}

func (o Conversion) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Conversion struct{}"
	}

	return strings.Join([]string{"Conversion", string(data)}, " ")
}
