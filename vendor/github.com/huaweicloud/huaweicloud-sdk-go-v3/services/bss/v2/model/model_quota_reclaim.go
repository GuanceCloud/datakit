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

type QuotaReclaim struct {
	// |参数名称：被回收的二级经销商代金券额度ID| |参数约束及描述：被回收的二级经销商代金券额度ID|
	QuotaId *string `json:"quota_id,omitempty"`
	// |参数名称：被回收的代金券的余额| |参数的约束及描述：被回收的代金券的余额|
	QuotaBalance float32 `json:"quota_balance,omitempty"`
}

func (o QuotaReclaim) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "QuotaReclaim struct{}"
	}

	return strings.Join([]string{"QuotaReclaim", string(data)}, " ")
}
