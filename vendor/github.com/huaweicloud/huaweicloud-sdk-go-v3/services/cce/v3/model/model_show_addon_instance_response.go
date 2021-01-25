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

// Response Object
type ShowAddonInstanceResponse struct {
	// API版本，固定值“v3”，该值不可修改。
	ApiVersion *string `json:"apiVersion,omitempty"`
	// API类型，固定值“Addon”，该值不可修改。
	Kind           *string              `json:"kind,omitempty"`
	Metadata       *Metadata            `json:"metadata,omitempty"`
	Spec           *InstanceSpec        `json:"spec,omitempty"`
	Status         *AddonInstanceStatus `json:"status,omitempty"`
	HttpStatusCode int                  `json:"-"`
}

func (o ShowAddonInstanceResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowAddonInstanceResponse struct{}"
	}

	return strings.Join([]string{"ShowAddonInstanceResponse", string(data)}, " ")
}
