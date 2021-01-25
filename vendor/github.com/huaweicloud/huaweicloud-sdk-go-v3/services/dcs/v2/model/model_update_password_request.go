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
type UpdatePasswordRequest struct {
	InstanceId string                      `json:"instance_id"`
	Body       *ModifyInstancePasswordBody `json:"body,omitempty"`
}

func (o UpdatePasswordRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdatePasswordRequest struct{}"
	}

	return strings.Join([]string{"UpdatePasswordRequest", string(data)}, " ")
}
