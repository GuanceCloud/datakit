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
type UpdatePortRequest struct {
	PortId string                 `json:"port_id"`
	Body   *UpdatePortRequestBody `json:"body,omitempty"`
}

func (o UpdatePortRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdatePortRequest struct{}"
	}

	return strings.Join([]string{"UpdatePortRequest", string(data)}, " ")
}
