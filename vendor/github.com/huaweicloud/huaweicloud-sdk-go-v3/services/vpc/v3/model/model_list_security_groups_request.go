/*
 * VPC
 *
 * VPC Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ListSecurityGroupsRequest struct {
	Limit               *int32    `json:"limit,omitempty"`
	Marker              *string   `json:"marker,omitempty"`
	Id                  *[]string `json:"id,omitempty"`
	Name                *[]string `json:"name,omitempty"`
	Description         *[]string `json:"description,omitempty"`
	EnterpriseProjectId *string   `json:"enterprise_project_id,omitempty"`
}

func (o ListSecurityGroupsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSecurityGroupsRequest struct{}"
	}

	return strings.Join([]string{"ListSecurityGroupsRequest", string(data)}, " ")
}
