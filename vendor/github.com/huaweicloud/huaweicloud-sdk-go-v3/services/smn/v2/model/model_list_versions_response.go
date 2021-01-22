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
type ListVersionsResponse struct {
	// 描述version相关对象的列表。
	Versions       *[]VersionItem `json:"versions,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o ListVersionsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListVersionsResponse struct{}"
	}

	return strings.Join([]string{"ListVersionsResponse", string(data)}, " ")
}
