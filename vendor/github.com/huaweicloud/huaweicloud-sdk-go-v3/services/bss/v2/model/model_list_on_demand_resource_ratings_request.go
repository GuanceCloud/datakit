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
type ListOnDemandResourceRatingsRequest struct {
	Body *RateOnDemandReq `json:"body,omitempty"`
}

func (o ListOnDemandResourceRatingsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListOnDemandResourceRatingsRequest struct{}"
	}

	return strings.Join([]string{"ListOnDemandResourceRatingsRequest", string(data)}, " ")
}
