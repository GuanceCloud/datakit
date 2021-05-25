/*
 * ECS
 *
 * ECS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ShowServerTagsResponse struct {
	// 标签列表
	Tags           *[]ServerTag `json:"tags,omitempty"`
	HttpStatusCode int          `json:"-"`
}

func (o ShowServerTagsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowServerTagsResponse struct{}"
	}

	return strings.Join([]string{"ShowServerTagsResponse", string(data)}, " ")
}
