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
type CreateSubnetRequest struct {
	Body *CreateSubnetRequestBody `json:"body,omitempty"`
}

func (o CreateSubnetRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateSubnetRequest struct{}"
	}

	return strings.Join([]string{"CreateSubnetRequest", string(data)}, " ")
}
