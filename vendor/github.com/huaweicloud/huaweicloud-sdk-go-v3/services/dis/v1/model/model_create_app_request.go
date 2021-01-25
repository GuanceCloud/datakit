/*
 * DIS
 *
 * DIS v1 API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type CreateAppRequest struct {
	Body *CreateAppRequest `json:"body,omitempty"`
}

func (o CreateAppRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateAppRequest struct{}"
	}

	return strings.Join([]string{"CreateAppRequest", string(data)}, " ")
}
