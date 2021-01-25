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

// Response Object
type CancelCustomerOrderResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CancelCustomerOrderResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CancelCustomerOrderResponse struct{}"
	}

	return strings.Join([]string{"CancelCustomerOrderResponse", string(data)}, " ")
}
