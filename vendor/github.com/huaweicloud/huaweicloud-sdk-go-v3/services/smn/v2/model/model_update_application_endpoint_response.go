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
type UpdateApplicationEndpointResponse struct {
	// 请求的唯一标识ID。
	RequestId      *string `json:"request_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o UpdateApplicationEndpointResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateApplicationEndpointResponse struct{}"
	}

	return strings.Join([]string{"UpdateApplicationEndpointResponse", string(data)}, " ")
}
