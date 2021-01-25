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

type OfficialWebsiteRatingResult struct {
	// |参数名称：官网价| |参数约束及描述：官网价|
	OfficialWebsiteAmount float32 `json:"official_website_amount,omitempty"`
	// |参数名称：度量单位标识1：元|
	MeasureId *int32 `json:"measure_id,omitempty"`
	// |参数名称：产品询价结果| |参数的约束及描述：产品询价结果|
	ProductRatingResults *[]PeriodProductOfficialRatingResult `json:"product_rating_results,omitempty"`
}

func (o OfficialWebsiteRatingResult) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "OfficialWebsiteRatingResult struct{}"
	}

	return strings.Join([]string{"OfficialWebsiteRatingResult", string(data)}, " ")
}
