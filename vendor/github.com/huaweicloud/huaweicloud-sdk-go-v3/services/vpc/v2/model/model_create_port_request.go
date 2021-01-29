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
type CreatePortRequest struct {
	Body *CreatePortRequestBody `json:"body,omitempty"`
}

func (o CreatePortRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreatePortRequest struct{}"
	}

	return strings.Join([]string{"CreatePortRequest", string(data)}, " ")
}
