/*
 * SMN
 *
 * SMN Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ListTopicsRequest struct {
	Offset *int32 `json:"offset,omitempty"`
	Limit  *int32 `json:"limit,omitempty"`
}

func (o ListTopicsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListTopicsRequest struct{}"
	}

	return strings.Join([]string{"ListTopicsRequest", string(data)}, " ")
}
