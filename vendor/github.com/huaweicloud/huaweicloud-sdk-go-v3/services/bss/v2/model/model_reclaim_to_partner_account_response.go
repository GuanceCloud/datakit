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
type ReclaimToPartnerAccountResponse struct {
	// |参数名称：事务流水ID，只有成功响应才会返回。| |参数约束及描述：事务流水ID，只有成功响应才会返回。|
	TransId        *string `json:"trans_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ReclaimToPartnerAccountResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ReclaimToPartnerAccountResponse struct{}"
	}

	return strings.Join([]string{"ReclaimToPartnerAccountResponse", string(data)}, " ")
}
