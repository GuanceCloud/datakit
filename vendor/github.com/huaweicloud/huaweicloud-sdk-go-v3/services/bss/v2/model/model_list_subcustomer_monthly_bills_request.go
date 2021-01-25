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
type ListSubcustomerMonthlyBillsRequest struct {
	CustomerId        *string `json:"customer_id,omitempty"`
	Cycle             string  `json:"cycle"`
	CloudServiceType  *string `json:"cloud_service_type,omitempty"`
	ChargeMode        string  `json:"charge_mode"`
	Offset            *int32  `json:"offset,omitempty"`
	Limit             *int32  `json:"limit,omitempty"`
	BillType          *string `json:"bill_type,omitempty"`
	IndirectPartnerId *string `json:"indirect_partner_id,omitempty"`
}

func (o ListSubcustomerMonthlyBillsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSubcustomerMonthlyBillsRequest struct{}"
	}

	return strings.Join([]string{"ListSubcustomerMonthlyBillsRequest", string(data)}, " ")
}
