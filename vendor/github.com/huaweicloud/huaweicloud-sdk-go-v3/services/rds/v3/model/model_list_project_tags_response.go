/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListProjectTagsResponse struct {
	// 标签列表，没有标签默认为空数组。
	Tags           *[]ProjectTagInfoResponse `json:"tags,omitempty"`
	HttpStatusCode int                       `json:"-"`
}

func (o ListProjectTagsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListProjectTagsResponse struct{}"
	}

	return strings.Join([]string{"ListProjectTagsResponse", string(data)}, " ")
}
