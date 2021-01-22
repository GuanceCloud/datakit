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
type UpdateConfigurationsRequest struct {
	InstanceId string                 `json:"instance_id"`
	Body       *ModifyRedisConfigBody `json:"body,omitempty"`
}

func (o UpdateConfigurationsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateConfigurationsRequest struct{}"
	}

	return strings.Join([]string{"UpdateConfigurationsRequest", string(data)}, " ")
}
