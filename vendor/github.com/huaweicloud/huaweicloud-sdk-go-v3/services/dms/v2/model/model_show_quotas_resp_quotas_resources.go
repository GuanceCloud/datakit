/*
 * DMS
 *
 * DMS Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ShowQuotasRespQuotasResources struct {
	// 配额名称。
	Type *string `json:"type,omitempty"`
	// 配额数量。
	Quota *int32 `json:"quota,omitempty"`
	// 已使用的数量。
	Used *int32 `json:"used,omitempty"`
	// 配额调整的最小值。
	Min *int32 `json:"min,omitempty"`
	// 配额调整的最大值。
	Max *int32 `json:"max,omitempty"`
}

func (o ShowQuotasRespQuotasResources) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowQuotasRespQuotasResources struct{}"
	}

	return strings.Join([]string{"ShowQuotasRespQuotasResources", string(data)}, " ")
}
