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
type ListRateOnPeriodDetailRequest struct {
	Body *RateOnPeriodReq `json:"body,omitempty"`
}

func (o ListRateOnPeriodDetailRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListRateOnPeriodDetailRequest struct{}"
	}

	return strings.Join([]string{"ListRateOnPeriodDetailRequest", string(data)}, " ")
}
