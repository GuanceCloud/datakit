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
type ListIndirectPartnersRequest struct {
	Body *QueryIndirectPartnersReq `json:"body,omitempty"`
}

func (o ListIndirectPartnersRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListIndirectPartnersRequest struct{}"
	}

	return strings.Join([]string{"ListIndirectPartnersRequest", string(data)}, " ")
}
