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

// Request Object
type PublishAppMessageRequest struct {
	EndpointUrn string                        `json:"endpoint_urn"`
	Body        *PublishAppMessageRequestBody `json:"body,omitempty"`
}

func (o PublishAppMessageRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PublishAppMessageRequest struct{}"
	}

	return strings.Join([]string{"PublishAppMessageRequest", string(data)}, " ")
}
