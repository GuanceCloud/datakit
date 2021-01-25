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

type QuotaSimpleInfo struct {
	// |参数名称：批量发放成功客户ID。| |参数约束及描述：批量发放成功客户ID。|
	Id string `json:"id"`
	// |参数名称：发放成功的券ID| |参数约束及描述：发放成功的券ID|
	QuotaId string `json:"quota_id"`
}

func (o QuotaSimpleInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "QuotaSimpleInfo struct{}"
	}

	return strings.Join([]string{"QuotaSimpleInfo", string(data)}, " ")
}
