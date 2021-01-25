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

type PeriodProductOfficialRatingResult struct {
	// |参数名称：ID标识，来源于请求中的ID| |参数约束及描述：ID标识，来源于请求中的ID|
	Id *string `json:"id,omitempty"`
	// |参数名称：产品ID| |参数约束及描述：产品ID|
	ProductId *string `json:"product_id,omitempty"`
	// |参数名称：官网价| |参数约束及描述：官网价|
	OfficialWebsiteAmount float32 `json:"official_website_amount,omitempty"`
	// |参数名称：度量单位标识1：元|
	MeasureId *int32 `json:"measure_id,omitempty"`
}

func (o PeriodProductOfficialRatingResult) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PeriodProductOfficialRatingResult struct{}"
	}

	return strings.Join([]string{"PeriodProductOfficialRatingResult", string(data)}, " ")
}
