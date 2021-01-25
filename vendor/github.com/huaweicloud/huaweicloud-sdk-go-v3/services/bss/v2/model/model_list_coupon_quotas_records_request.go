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
type ListCouponQuotasRecordsRequest struct {
	IndirectPartnerId  *string `json:"indirect_partner_id,omitempty"`
	QuotaId            *string `json:"quota_id,omitempty"`
	OperationTimeBegin *string `json:"operation_time_begin,omitempty"`
	OperationTimeEnd   *string `json:"operation_time_end,omitempty"`
	ParentQuotaId      *string `json:"parent_quota_id,omitempty"`
	OperationType      *string `json:"operation_type,omitempty"`
	Offset             *int32  `json:"offset,omitempty"`
	Limit              *int32  `json:"limit,omitempty"`
}

func (o ListCouponQuotasRecordsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListCouponQuotasRecordsRequest struct{}"
	}

	return strings.Join([]string{"ListCouponQuotasRecordsRequest", string(data)}, " ")
}
