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
type CreatePrivateipRequest struct {
	Body *CreatePrivateipRequestBody `json:"body,omitempty"`
}

func (o CreatePrivateipRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreatePrivateipRequest struct{}"
	}

	return strings.Join([]string{"CreatePrivateipRequest", string(data)}, " ")
}
