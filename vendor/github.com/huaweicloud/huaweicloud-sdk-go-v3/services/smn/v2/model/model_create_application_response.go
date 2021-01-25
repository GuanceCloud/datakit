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
type CreateApplicationResponse struct {
	// 请求的唯一标识ID。
	RequestId *string `json:"request_id,omitempty"`
	// Application的唯一资源标识。
	ApplicationUrn *string `json:"application_urn,omitempty"`
	// Application资源的ID。
	ApplicationId  *string `json:"application_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CreateApplicationResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateApplicationResponse struct{}"
	}

	return strings.Join([]string{"CreateApplicationResponse", string(data)}, " ")
}
