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
type CreateApplicationRequest struct {
	Body *CreateApplicationRequestBody `json:"body,omitempty"`
}

func (o CreateApplicationRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateApplicationRequest struct{}"
	}

	return strings.Join([]string{"CreateApplicationRequest", string(data)}, " ")
}
