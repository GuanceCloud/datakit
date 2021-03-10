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
type CreateTrackerConfigRequest struct {
	Body *TrackerConfigBody `json:"body,omitempty"`
}

func (o CreateTrackerConfigRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateTrackerConfigRequest struct{}"
	}

	return strings.Join([]string{"CreateTrackerConfigRequest", string(data)}, " ")
}
