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
type CreateAppV3Request struct {
	Body *CreateAppRequest `json:"body,omitempty"`
}

func (o CreateAppV3Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateAppV3Request struct{}"
	}

	return strings.Join([]string{"CreateAppV3Request", string(data)}, " ")
}
