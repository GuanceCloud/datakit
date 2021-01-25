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
type ShowRefundOrderDetailsResponse struct {
	// |参数名称：总记录数。| |参数的约束及描述：总记录数。|
	TotalCount *int32 `json:"total_count,omitempty"`
	// |参数名称：资源信息列表。具体请参见表2 OrderRefundInfoV2。| |参数约束以及描述：资源信息列表。具体请参见表2 OrderRefundInfoV2。|
	RefundInfos    *[]OrderRefundInfoV2 `json:"refund_infos,omitempty"`
	HttpStatusCode int                  `json:"-"`
}

func (o ShowRefundOrderDetailsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowRefundOrderDetailsResponse struct{}"
	}

	return strings.Join([]string{"ShowRefundOrderDetailsResponse", string(data)}, " ")
}
