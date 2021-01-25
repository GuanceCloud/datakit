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

type Resources struct {
	// 资源的计数单位。 - 当type为instance时，无单位。 - 当type为ram时，单位为GB。
	Unit *string `json:"unit,omitempty"`
	// - 当type为instance时，表示可申请实例配额的最小值。 - 当type为ram时，表示可申请内存配额的最小值。
	Min *int32 `json:"min,omitempty"`
	// - 当type为instance时，表示可申请实例配额的最大值。 - 当type为ram时，表示可申请内存配额的最大值。
	Max *int32 `json:"max,omitempty"`
	// 可以创建的实例最大数和总内存的配额限制。
	Quota *int32 `json:"quota,omitempty"`
	// 已创建的实例个数和已使用的内存配额。
	Used *int32 `json:"used,omitempty"`
	// 支持instance、ram两种。 - instance表示实例配额。 - ram表示内存配额。
	Type *string `json:"type,omitempty"`
}

func (o Resources) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Resources struct{}"
	}

	return strings.Join([]string{"Resources", string(data)}, " ")
}
