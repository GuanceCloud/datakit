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
type ListPartnerCouponsRecordResponse struct {
	// |参数名称：查询记录总数。| |参数的约束及描述：查询记录总数。|
	TotalCount *int32 `json:"total_count,omitempty"`
	// |参数名称：日志记录列表。具体请参见表2 CouponRecordV2。| |参数约束以及描述：日志记录列表。具体请参见表2 CouponRecordV2。|
	Records        *[]CouponRecordV2 `json:"records,omitempty"`
	HttpStatusCode int               `json:"-"`
}

func (o ListPartnerCouponsRecordResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPartnerCouponsRecordResponse struct{}"
	}

	return strings.Join([]string{"ListPartnerCouponsRecordResponse", string(data)}, " ")
}
