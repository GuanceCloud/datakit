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
type ListPartnerBalancesRequest struct {
	IndirectPartnerId *string `json:"indirect_partner_id,omitempty"`
}

func (o ListPartnerBalancesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPartnerBalancesRequest struct{}"
	}

	return strings.Join([]string{"ListPartnerBalancesRequest", string(data)}, " ")
}
