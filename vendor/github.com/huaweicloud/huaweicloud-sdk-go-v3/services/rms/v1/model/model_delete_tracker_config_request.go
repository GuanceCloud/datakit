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

// Request Object
type DeleteTrackerConfigRequest struct {
}

func (o DeleteTrackerConfigRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteTrackerConfigRequest struct{}"
	}

	return strings.Join([]string{"DeleteTrackerConfigRequest", string(data)}, " ")
}
