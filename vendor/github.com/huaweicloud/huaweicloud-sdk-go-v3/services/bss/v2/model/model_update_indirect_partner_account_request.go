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
type UpdateIndirectPartnerAccountRequest struct {
	Body *AdjustToIndirectPartnerReq `json:"body,omitempty"`
}

func (o UpdateIndirectPartnerAccountRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateIndirectPartnerAccountRequest struct{}"
	}

	return strings.Join([]string{"UpdateIndirectPartnerAccountRequest", string(data)}, " ")
}
