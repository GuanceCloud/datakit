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
type CheckUserIdentityRequest struct {
	Body *CheckSubcustomerUserReq `json:"body,omitempty"`
}

func (o CheckUserIdentityRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CheckUserIdentityRequest struct{}"
	}

	return strings.Join([]string{"CheckUserIdentityRequest", string(data)}, " ")
}
