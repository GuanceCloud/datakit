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

// Response Object
type ListProjectTagsResponse struct {
	// 资源标签列表。
	Tags           *[]ResourceTags `json:"tags,omitempty"`
	HttpStatusCode int             `json:"-"`
}

func (o ListProjectTagsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListProjectTagsResponse struct{}"
	}

	return strings.Join([]string{"ListProjectTagsResponse", string(data)}, " ")
}
