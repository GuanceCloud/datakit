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
type UpdateIndirectPartnerAccountResponse struct {
	// |参数名称：事务流水ID，只有成功响应才会返回。| |参数约束及描述：非必填|
	TransferId     *string `json:"transfer_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o UpdateIndirectPartnerAccountResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateIndirectPartnerAccountResponse struct{}"
	}

	return strings.Join([]string{"UpdateIndirectPartnerAccountResponse", string(data)}, " ")
}
