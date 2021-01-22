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
type UpdatePeriodToOnDemandRequest struct {
	Body *PeriodToOnDemandReq `json:"body,omitempty"`
}

func (o UpdatePeriodToOnDemandRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdatePeriodToOnDemandRequest struct{}"
	}

	return strings.Join([]string{"UpdatePeriodToOnDemandRequest", string(data)}, " ")
}
