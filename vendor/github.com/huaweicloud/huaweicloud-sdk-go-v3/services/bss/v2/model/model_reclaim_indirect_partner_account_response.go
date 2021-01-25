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
type ReclaimIndirectPartnerAccountResponse struct {
	// |参数名称：回收流水| |参数约束及描述：回收流水|
	TransId        *string `json:"trans_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ReclaimIndirectPartnerAccountResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ReclaimIndirectPartnerAccountResponse struct{}"
	}

	return strings.Join([]string{"ReclaimIndirectPartnerAccountResponse", string(data)}, " ")
}
