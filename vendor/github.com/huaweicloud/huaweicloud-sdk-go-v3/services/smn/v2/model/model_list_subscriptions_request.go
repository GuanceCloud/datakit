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
type ListSubscriptionsRequest struct {
	Offset *int32 `json:"offset,omitempty"`
	Limit  *int32 `json:"limit,omitempty"`
}

func (o ListSubscriptionsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSubscriptionsRequest struct{}"
	}

	return strings.Join([]string{"ListSubscriptionsRequest", string(data)}, " ")
}
