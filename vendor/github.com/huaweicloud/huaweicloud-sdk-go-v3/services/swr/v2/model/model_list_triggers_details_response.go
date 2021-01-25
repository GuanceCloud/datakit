/*
 * SWR
 *
 * SWR API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListTriggersDetailsResponse struct {
	// 触发器列表
	Body           *[]Trigger `json:"body,omitempty"`
	HttpStatusCode int        `json:"-"`
}

func (o ListTriggersDetailsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListTriggersDetailsResponse struct{}"
	}

	return strings.Join([]string{"ListTriggersDetailsResponse", string(data)}, " ")
}
