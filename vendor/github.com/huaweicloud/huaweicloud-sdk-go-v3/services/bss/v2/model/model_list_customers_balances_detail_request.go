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
type ListCustomersBalancesDetailRequest struct {
	Body *QueryCustomersBalancesReq `json:"body,omitempty"`
}

func (o ListCustomersBalancesDetailRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListCustomersBalancesDetailRequest struct{}"
	}

	return strings.Join([]string{"ListCustomersBalancesDetailRequest", string(data)}, " ")
}
