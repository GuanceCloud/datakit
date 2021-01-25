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
type CreateSubEnterpriseAccountRequest struct {
	Body *CreateSubCustomerReqV2 `json:"body,omitempty"`
}

func (o CreateSubEnterpriseAccountRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateSubEnterpriseAccountRequest struct{}"
	}

	return strings.Join([]string{"CreateSubEnterpriseAccountRequest", string(data)}, " ")
}
