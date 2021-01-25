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
type CreateApplicationEndpointResponse struct {
	// 请求的唯一标识ID。
	RequestId *string `json:"request_id,omitempty"`
	// Endpoint的唯一资源标识。
	EndpointUrn    *string `json:"endpoint_urn,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CreateApplicationEndpointResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateApplicationEndpointResponse struct{}"
	}

	return strings.Join([]string{"CreateApplicationEndpointResponse", string(data)}, " ")
}
