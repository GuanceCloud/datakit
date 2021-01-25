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
type ShowCustomerAccountBalancesRequest struct {
}

func (o ShowCustomerAccountBalancesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowCustomerAccountBalancesRequest struct{}"
	}

	return strings.Join([]string{"ShowCustomerAccountBalancesRequest", string(data)}, " ")
}
