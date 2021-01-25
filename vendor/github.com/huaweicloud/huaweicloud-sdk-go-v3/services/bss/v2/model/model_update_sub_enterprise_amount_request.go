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
type UpdateSubEnterpriseAmountRequest struct {
	Body *TransferEnterpriseMultiAccountReq `json:"body,omitempty"`
}

func (o UpdateSubEnterpriseAmountRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateSubEnterpriseAmountRequest struct{}"
	}

	return strings.Join([]string{"UpdateSubEnterpriseAmountRequest", string(data)}, " ")
}
