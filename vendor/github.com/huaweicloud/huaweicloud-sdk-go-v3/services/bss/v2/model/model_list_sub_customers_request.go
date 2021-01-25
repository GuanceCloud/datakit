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
type ListSubCustomersRequest struct {
	Body *QuerySubCustomerListReq `json:"body,omitempty"`
}

func (o ListSubCustomersRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSubCustomersRequest struct{}"
	}

	return strings.Join([]string{"ListSubCustomersRequest", string(data)}, " ")
}
