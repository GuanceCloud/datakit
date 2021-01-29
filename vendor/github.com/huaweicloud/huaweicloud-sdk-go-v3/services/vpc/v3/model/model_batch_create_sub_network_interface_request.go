/*
 * VPC
 *
 * VPC Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type BatchCreateSubNetworkInterfaceRequest struct {
	Body *BatchCreateSubNetworkInterfaceRequestBody `json:"body,omitempty"`
}

func (o BatchCreateSubNetworkInterfaceRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchCreateSubNetworkInterfaceRequest struct{}"
	}

	return strings.Join([]string{"BatchCreateSubNetworkInterfaceRequest", string(data)}, " ")
}
