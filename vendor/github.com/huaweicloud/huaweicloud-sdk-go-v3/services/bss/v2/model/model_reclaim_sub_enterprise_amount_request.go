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
type ReclaimSubEnterpriseAmountRequest struct {
	Body *RetrieveEnterpriseMultiAccountReq `json:"body,omitempty"`
}

func (o ReclaimSubEnterpriseAmountRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ReclaimSubEnterpriseAmountRequest struct{}"
	}

	return strings.Join([]string{"ReclaimSubEnterpriseAmountRequest", string(data)}, " ")
}
