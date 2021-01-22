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
type ReclaimIndirectPartnerAccountRequest struct {
	Body *ReclaimIndirectPartnerAccountReq `json:"body,omitempty"`
}

func (o ReclaimIndirectPartnerAccountRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ReclaimIndirectPartnerAccountRequest struct{}"
	}

	return strings.Join([]string{"ReclaimIndirectPartnerAccountRequest", string(data)}, " ")
}
