/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 配额信息。
type QueryTenantQuotaRespQuotas struct {
	// 配额列表。
	Resources *[]Resources `json:"resources,omitempty"`
}

func (o QueryTenantQuotaRespQuotas) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "QueryTenantQuotaRespQuotas struct{}"
	}

	return strings.Join([]string{"QueryTenantQuotaRespQuotas", string(data)}, " ")
}
