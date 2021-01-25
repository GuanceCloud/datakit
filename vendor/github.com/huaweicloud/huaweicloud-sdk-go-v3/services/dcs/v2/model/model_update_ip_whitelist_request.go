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
type UpdateIpWhitelistRequest struct {
	InstanceId string                 `json:"instance_id"`
	Body       *ModifyIpWhitelistBody `json:"body,omitempty"`
}

func (o UpdateIpWhitelistRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateIpWhitelistRequest struct{}"
	}

	return strings.Join([]string{"UpdateIpWhitelistRequest", string(data)}, " ")
}
