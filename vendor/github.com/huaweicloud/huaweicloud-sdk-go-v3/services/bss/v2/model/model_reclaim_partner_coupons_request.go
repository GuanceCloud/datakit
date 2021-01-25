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
type ReclaimPartnerCouponsRequest struct {
	Body *ReclaimPartnerCouponsReq `json:"body,omitempty"`
}

func (o ReclaimPartnerCouponsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ReclaimPartnerCouponsRequest struct{}"
	}

	return strings.Join([]string{"ReclaimPartnerCouponsRequest", string(data)}, " ")
}
