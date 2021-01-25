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
type ReclaimCouponQuotasResponse struct {
	// |参数名称：响应信息| |参数约束以及描述：响应信息|
	ErrorDetails *[]ErrorDetail `json:"error_details,omitempty"`
	// |参数名称：响应信息| |参数约束以及描述：响应信息|
	SimpleQuotaInfos *[]QuotaReclaim `json:"simple_quota_infos,omitempty"`
	HttpStatusCode   int             `json:"-"`
}

func (o ReclaimCouponQuotasResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ReclaimCouponQuotasResponse struct{}"
	}

	return strings.Join([]string{"ReclaimCouponQuotasResponse", string(data)}, " ")
}
