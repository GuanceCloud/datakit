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
type ListCustomerOrdersRequest struct {
	OrderId           *string `json:"order_id,omitempty"`
	CustomerId        *string `json:"customer_id,omitempty"`
	CreateTimeBegin   *string `json:"create_time_begin,omitempty"`
	CreateTimeEnd     *string `json:"create_time_end,omitempty"`
	ServiceTypeCode   *string `json:"service_type_code,omitempty"`
	Status            *int32  `json:"status,omitempty"`
	OrderType         *string `json:"order_type,omitempty"`
	Limit             *int32  `json:"limit,omitempty"`
	Offset            *int32  `json:"offset,omitempty"`
	OrderBy           *string `json:"order_by,omitempty"`
	PaymentTimeBegin  *string `json:"payment_time_begin,omitempty"`
	PaymentTimeEnd    *string `json:"payment_time_end,omitempty"`
	IndirectPartnerId *string `json:"indirect_partner_id,omitempty"`
}

func (o ListCustomerOrdersRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListCustomerOrdersRequest struct{}"
	}

	return strings.Join([]string{"ListCustomerOrdersRequest", string(data)}, " ")
}
