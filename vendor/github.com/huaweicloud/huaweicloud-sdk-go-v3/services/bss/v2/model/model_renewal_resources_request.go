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
type RenewalResourcesRequest struct {
	Body *RenewalResourcesReq `json:"body,omitempty"`
}

func (o RenewalResourcesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RenewalResourcesRequest struct{}"
	}

	return strings.Join([]string{"RenewalResourcesRequest", string(data)}, " ")
}
