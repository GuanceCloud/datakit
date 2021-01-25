/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type BatchDeleteInstancesRequest struct {
	AllFailure *bool            `json:"all_failure,omitempty"`
	Body       *BatchDeleteBody `json:"body,omitempty"`
}

func (o BatchDeleteInstancesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchDeleteInstancesRequest struct{}"
	}

	return strings.Join([]string{"BatchDeleteInstancesRequest", string(data)}, " ")
}
