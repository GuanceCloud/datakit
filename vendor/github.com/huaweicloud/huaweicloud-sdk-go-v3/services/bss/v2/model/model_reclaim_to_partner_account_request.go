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
type ReclaimToPartnerAccountRequest struct {
	Body *ReclaimToPartnerAccountBalancesReq `json:"body,omitempty"`
}

func (o ReclaimToPartnerAccountRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ReclaimToPartnerAccountRequest struct{}"
	}

	return strings.Join([]string{"ReclaimToPartnerAccountRequest", string(data)}, " ")
}
