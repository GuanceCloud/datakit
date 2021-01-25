/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type CreateBareMetalServersResponse struct {
	// 订单ID
	OrderId *string `json:"order_id,omitempty"`
	// 任务ID
	JobId          *string `json:"job_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CreateBareMetalServersResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateBareMetalServersResponse struct{}"
	}

	return strings.Join([]string{"CreateBareMetalServersResponse", string(data)}, " ")
}
