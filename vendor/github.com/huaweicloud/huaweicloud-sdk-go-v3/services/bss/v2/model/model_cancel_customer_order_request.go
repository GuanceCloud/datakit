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
type CancelCustomerOrderRequest struct {
	Body *CancelCustomerOrderReq `json:"body,omitempty"`
}

func (o CancelCustomerOrderRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CancelCustomerOrderRequest struct{}"
	}

	return strings.Join([]string{"CancelCustomerOrderRequest", string(data)}, " ")
}
