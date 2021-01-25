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
type ListIssuedCouponQuotasRequest struct {
	QuotaId           *string `json:"quota_id,omitempty"`
	IndirectPartnerId *string `json:"indirect_partner_id,omitempty"`
	ParentQuotaId     *string `json:"parent_quota_id,omitempty"`
	Offset            *int32  `json:"offset,omitempty"`
	Limit             *int32  `json:"limit,omitempty"`
}

func (o ListIssuedCouponQuotasRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListIssuedCouponQuotasRequest struct{}"
	}

	return strings.Join([]string{"ListIssuedCouponQuotasRequest", string(data)}, " ")
}
