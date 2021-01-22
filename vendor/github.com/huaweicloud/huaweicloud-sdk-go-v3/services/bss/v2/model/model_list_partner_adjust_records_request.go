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
type ListPartnerAdjustRecordsRequest struct {
	CustomerId         *string `json:"customer_id,omitempty"`
	OperationType      *string `json:"operation_type,omitempty"`
	OperationTimeBegin *string `json:"operation_time_begin,omitempty"`
	OperationTimeEnd   *string `json:"operation_time_end,omitempty"`
	TransId            *string `json:"trans_id,omitempty"`
	Offset             *int32  `json:"offset,omitempty"`
	Limit              *int32  `json:"limit,omitempty"`
	IndirectPartnerId  *string `json:"indirect_partner_id,omitempty"`
}

func (o ListPartnerAdjustRecordsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPartnerAdjustRecordsRequest struct{}"
	}

	return strings.Join([]string{"ListPartnerAdjustRecordsRequest", string(data)}, " ")
}
