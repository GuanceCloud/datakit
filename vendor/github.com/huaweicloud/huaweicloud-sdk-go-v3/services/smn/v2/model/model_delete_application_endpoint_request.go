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
type DeleteApplicationEndpointRequest struct {
	EndpointUrn string `json:"endpoint_urn"`
}

func (o DeleteApplicationEndpointRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteApplicationEndpointRequest struct{}"
	}

	return strings.Join([]string{"DeleteApplicationEndpointRequest", string(data)}, " ")
}
