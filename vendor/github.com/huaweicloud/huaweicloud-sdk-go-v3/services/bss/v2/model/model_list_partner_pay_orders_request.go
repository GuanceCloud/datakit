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
type ListPartnerPayOrdersRequest struct {
	OrderId           *string `json:"order_id,omitempty"`
	CustomerId        *string `json:"customer_id,omitempty"`
	Limit             *int32  `json:"limit,omitempty"`
	Offset            *int32  `json:"offset,omitempty"`
	Status            *int32  `json:"status,omitempty"`
	IndirectPartnerId *string `json:"indirect_partner_id,omitempty"`
}

func (o ListPartnerPayOrdersRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPartnerPayOrdersRequest struct{}"
	}

	return strings.Join([]string{"ListPartnerPayOrdersRequest", string(data)}, " ")
}
