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
type ListQuotaCouponsResponse struct {
	// |参数名称：查询总数。| |参数的约束及描述：查询总数。|
	TotalCount *int32 `json:"total_count,omitempty"`
	// |参数名称：额度记录列表。具体请参见表2 IssuedCouponQuota。| |参数约束以及描述：额度记录列表。具体请参见表2 IssuedCouponQuota。|
	Quotas         *[]CouponQuotaV2 `json:"quotas,omitempty"`
	HttpStatusCode int              `json:"-"`
}

func (o ListQuotaCouponsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListQuotaCouponsResponse struct{}"
	}

	return strings.Join([]string{"ListQuotaCouponsResponse", string(data)}, " ")
}
