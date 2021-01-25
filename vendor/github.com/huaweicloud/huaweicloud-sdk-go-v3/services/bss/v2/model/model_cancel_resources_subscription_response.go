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
type CancelResourcesSubscriptionResponse struct {
	// |参数名称：退订资源生成的订单ID的列表。| |参数约束以及描述：续订资源生成的订单ID的列表。|
	OrderIds       *[]string `json:"order_ids,omitempty"`
	HttpStatusCode int       `json:"-"`
}

func (o CancelResourcesSubscriptionResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CancelResourcesSubscriptionResponse struct{}"
	}

	return strings.Join([]string{"CancelResourcesSubscriptionResponse", string(data)}, " ")
}
