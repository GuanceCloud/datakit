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
type ShowJobResponse struct {
	// API版本，固定值“v3”，该值不可修改。
	ApiVersion *string `json:"apiVersion,omitempty"`
	// API类型，固定值“Job”，该值不可修改。
	Kind           *string         `json:"kind,omitempty"`
	Metadata       *CceJobMetadata `json:"metadata,omitempty"`
	Spec           *CceJobSpec     `json:"spec,omitempty"`
	Status         *CceJobStatus   `json:"status,omitempty"`
	HttpStatusCode int             `json:"-"`
}

func (o ShowJobResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowJobResponse struct{}"
	}

	return strings.Join([]string{"ShowJobResponse", string(data)}, " ")
}
