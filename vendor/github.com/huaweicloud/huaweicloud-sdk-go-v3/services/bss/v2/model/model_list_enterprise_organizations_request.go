/*
 * BSS
 *
 * Business Support System API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ListEnterpriseOrganizationsRequest struct {
	RecursiveQuery *int32  `json:"recursive_query,omitempty"`
	ParentId       *string `json:"parent_id,omitempty"`
}

func (o ListEnterpriseOrganizationsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListEnterpriseOrganizationsRequest struct{}"
	}

	return strings.Join([]string{"ListEnterpriseOrganizationsRequest", string(data)}, " ")
}
