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

// Request Object
type ShowRefundOrderDetailsRequest struct {
	OrderId string `json:"order_id"`
}

func (o ShowRefundOrderDetailsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowRefundOrderDetailsRequest struct{}"
	}

	return strings.Join([]string{"ShowRefundOrderDetailsRequest", string(data)}, " ")
}
