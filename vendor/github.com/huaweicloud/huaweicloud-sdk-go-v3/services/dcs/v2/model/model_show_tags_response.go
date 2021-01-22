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

// Response Object
type ShowTagsResponse struct {
	// 标签列表。
	Tags           *[]ResourceTag `json:"tags,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o ShowTagsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowTagsResponse struct{}"
	}

	return strings.Join([]string{"ShowTagsResponse", string(data)}, " ")
}
