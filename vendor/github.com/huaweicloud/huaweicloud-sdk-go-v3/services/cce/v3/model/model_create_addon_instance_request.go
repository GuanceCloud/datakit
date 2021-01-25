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
type CreateAddonInstanceRequest struct {
	ContentType string           `json:"Content-Type"`
	Body        *InstanceRequest `json:"body,omitempty"`
}

func (o CreateAddonInstanceRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateAddonInstanceRequest struct{}"
	}

	return strings.Join([]string{"CreateAddonInstanceRequest", string(data)}, " ")
}
