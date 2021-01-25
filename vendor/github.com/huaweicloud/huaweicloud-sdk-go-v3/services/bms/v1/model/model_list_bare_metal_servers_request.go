/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ListBareMetalServersRequest struct {
	Flavor              *string `json:"flavor,omitempty"`
	Name                *string `json:"name,omitempty"`
	Status              *string `json:"status,omitempty"`
	Limit               *int32  `json:"limit,omitempty"`
	Offset              *int32  `json:"offset,omitempty"`
	Tags                *string `json:"tags,omitempty"`
	ReservationId       *string `json:"reservation_id,omitempty"`
	Detail              *string `json:"detail,omitempty"`
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
}

func (o ListBareMetalServersRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListBareMetalServersRequest struct{}"
	}

	return strings.Join([]string{"ListBareMetalServersRequest", string(data)}, " ")
}
