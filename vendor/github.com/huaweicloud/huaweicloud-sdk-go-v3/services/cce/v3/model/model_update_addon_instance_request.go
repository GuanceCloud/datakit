/*
 * CCE
 *
 * CCE开放API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type UpdateAddonInstanceRequest struct {
	Id          string           `json:"id"`
	ContentType string           `json:"Content-Type"`
	Body        *InstanceRequest `json:"body,omitempty"`
}

func (o UpdateAddonInstanceRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateAddonInstanceRequest struct{}"
	}

	return strings.Join([]string{"UpdateAddonInstanceRequest", string(data)}, " ")
}
