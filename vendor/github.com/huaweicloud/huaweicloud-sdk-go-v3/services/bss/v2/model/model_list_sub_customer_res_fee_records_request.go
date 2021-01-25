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
type ListSubCustomerResFeeRecordsRequest struct {
	CustomerId        string  `json:"customer_id"`
	Cycle             string  `json:"cycle"`
	CloudServiceType  *string `json:"cloud_service_type,omitempty"`
	Region            *string `json:"region,omitempty"`
	ChargeMode        string  `json:"charge_mode"`
	BillType          *int32  `json:"bill_type,omitempty"`
	Offset            *int32  `json:"offset,omitempty"`
	Limit             *int32  `json:"limit,omitempty"`
	ResourceId        *string `json:"resource_id,omitempty"`
	IncludeZeroRecord *bool   `json:"include_zero_record,omitempty"`
	IndirectPartnerId *string `json:"indirect_partner_id,omitempty"`
}

func (o ListSubCustomerResFeeRecordsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSubCustomerResFeeRecordsRequest struct{}"
	}

	return strings.Join([]string{"ListSubCustomerResFeeRecordsRequest", string(data)}, " ")
}
