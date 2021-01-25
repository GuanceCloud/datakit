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
type ListPolicyStatesByDomainIdRequest struct {
	ComplianceState *string `json:"compliance_state,omitempty"`
	ResourceId      *string `json:"resource_id,omitempty"`
	ResourceName    *string `json:"resource_name,omitempty"`
	Limit           *int32  `json:"limit,omitempty"`
	Marker          *string `json:"marker,omitempty"`
}

func (o ListPolicyStatesByDomainIdRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPolicyStatesByDomainIdRequest struct{}"
	}

	return strings.Join([]string{"ListPolicyStatesByDomainIdRequest", string(data)}, " ")
}
