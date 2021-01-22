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

// Response Object
type ListMeasureUnitsResponse struct {
	// |参数名称：度量信息| |参数约束以及描述：度量信息|
	MeasureUnits   *[]MeasureUnitRest `json:"measure_units,omitempty"`
	HttpStatusCode int                `json:"-"`
}

func (o ListMeasureUnitsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListMeasureUnitsResponse struct{}"
	}

	return strings.Join([]string{"ListMeasureUnitsResponse", string(data)}, " ")
}
