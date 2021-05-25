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
type DeleteTrackerConfigResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteTrackerConfigResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteTrackerConfigResponse struct{}"
	}

	return strings.Join([]string{"DeleteTrackerConfigResponse", string(data)}, " ")
}
