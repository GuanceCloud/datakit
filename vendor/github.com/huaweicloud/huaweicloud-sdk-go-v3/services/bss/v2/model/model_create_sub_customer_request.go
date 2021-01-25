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
type CreateSubCustomerRequest struct {
	Body *CreateCustomerV2Req `json:"body,omitempty"`
}

func (o CreateSubCustomerRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateSubCustomerRequest struct{}"
	}

	return strings.Join([]string{"CreateSubCustomerRequest", string(data)}, " ")
}
