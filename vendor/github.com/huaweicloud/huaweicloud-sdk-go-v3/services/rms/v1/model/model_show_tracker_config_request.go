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
type ShowTrackerConfigRequest struct {
}

func (o ShowTrackerConfigRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowTrackerConfigRequest struct{}"
	}

	return strings.Join([]string{"ShowTrackerConfigRequest", string(data)}, " ")
}
