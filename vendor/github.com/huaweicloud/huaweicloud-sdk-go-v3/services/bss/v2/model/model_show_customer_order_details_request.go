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
type ShowCustomerOrderDetailsRequest struct {
	OrderId           string  `json:"order_id"`
	Limit             *int32  `json:"limit,omitempty"`
	Offset            *int32  `json:"offset,omitempty"`
	IndirectPartnerId *string `json:"indirect_partner_id,omitempty"`
}

func (o ShowCustomerOrderDetailsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowCustomerOrderDetailsRequest struct{}"
	}

	return strings.Join([]string{"ShowCustomerOrderDetailsRequest", string(data)}, " ")
}
