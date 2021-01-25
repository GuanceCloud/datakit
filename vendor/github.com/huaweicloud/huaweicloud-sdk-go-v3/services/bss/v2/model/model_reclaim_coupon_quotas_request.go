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

// Request Object
type ReclaimCouponQuotasRequest struct {
	Body *ReclaimCouponQuotasReq `json:"body,omitempty"`
}

func (o ReclaimCouponQuotasRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ReclaimCouponQuotasRequest struct{}"
	}

	return strings.Join([]string{"ReclaimCouponQuotasRequest", string(data)}, " ")
}
