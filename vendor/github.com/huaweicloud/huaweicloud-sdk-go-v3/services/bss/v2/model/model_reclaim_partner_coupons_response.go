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

// Response Object
type ReclaimPartnerCouponsResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o ReclaimPartnerCouponsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ReclaimPartnerCouponsResponse struct{}"
	}

	return strings.Join([]string{"ReclaimPartnerCouponsResponse", string(data)}, " ")
}
