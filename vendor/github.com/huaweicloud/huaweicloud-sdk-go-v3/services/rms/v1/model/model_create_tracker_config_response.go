/*
 * RMS
 *
 * Resource Manager Api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type CreateTrackerConfigResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CreateTrackerConfigResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateTrackerConfigResponse struct{}"
	}

	return strings.Join([]string{"CreateTrackerConfigResponse", string(data)}, " ")
}
