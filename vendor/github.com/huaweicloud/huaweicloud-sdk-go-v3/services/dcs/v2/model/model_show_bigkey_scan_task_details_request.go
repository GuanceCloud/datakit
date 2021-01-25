/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ShowBigkeyScanTaskDetailsRequest struct {
	InstanceId string `json:"instance_id"`
	BigkeyId   string `json:"bigkey_id"`
}

func (o ShowBigkeyScanTaskDetailsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowBigkeyScanTaskDetailsRequest struct{}"
	}

	return strings.Join([]string{"ShowBigkeyScanTaskDetailsRequest", string(data)}, " ")
}
