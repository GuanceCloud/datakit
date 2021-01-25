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
type ListApplicationAttributesResponse struct {
	// 请求的唯一标识ID。
	RequestId *string `json:"request_id,omitempty"`
	// Application的唯一标识ID。
	ApplicationId  *string                                          `json:"application_id,omitempty"`
	Attributes     *ListApplicationAttributesResponseBodyAttributes `json:"attributes,omitempty"`
	HttpStatusCode int                                              `json:"-"`
}

func (o ListApplicationAttributesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListApplicationAttributesResponse struct{}"
	}

	return strings.Join([]string{"ListApplicationAttributesResponse", string(data)}, " ")
}
