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
type UpdateCouponQuotasResponse struct {
	// |参数名称：错误的客户列表和错误信息| |参数约束以及描述：错误的客户列表和错误信息|
	ErrorDetails *[]ErrorDetail `json:"error_details,omitempty"`
	// |参数名称：成功的客户ID和对应的券ID列表| |参数约束以及描述：成功的客户ID和对应的券ID列表|
	SimpleQuotaInfos *[]QuotaSimpleInfo `json:"simple_quota_infos,omitempty"`
	HttpStatusCode   int                `json:"-"`
}

func (o UpdateCouponQuotasResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateCouponQuotasResponse struct{}"
	}

	return strings.Join([]string{"UpdateCouponQuotasResponse", string(data)}, " ")
}
