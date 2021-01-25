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
type ListPartnerCouponsRecordRequest struct {
	OperationTypes     *[]string `json:"operation_types,omitempty"`
	QuotaId            *string   `json:"quota_id,omitempty"`
	QuotaType          *int32    `json:"quota_type,omitempty"`
	CouponIds          *[]string `json:"coupon_ids,omitempty"`
	CustomerId         *string   `json:"customer_id,omitempty"`
	OperationTimeBegin *string   `json:"operation_time_begin,omitempty"`
	OperationTimeEnd   *string   `json:"operation_time_end,omitempty"`
	Result             *string   `json:"result,omitempty"`
	Offset             *int32    `json:"offset,omitempty"`
	Limit              *int32    `json:"limit,omitempty"`
	IndirectPartnerId  *string   `json:"indirect_partner_id,omitempty"`
}

func (o ListPartnerCouponsRecordRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPartnerCouponsRecordRequest struct{}"
	}

	return strings.Join([]string{"ListPartnerCouponsRecordRequest", string(data)}, " ")
}
