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
type UpdatePeriodToOnDemandResponse struct {
	// |参数名称：返回数据| |参数约束以及描述：返回数据 HTTP 200的时候返回该字段，部分成功部分失败的时候返回的失败记录，如果全成功，该记录为空|
	ErrorDetails   *[]ErrorDetail `json:"error_details,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o UpdatePeriodToOnDemandResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdatePeriodToOnDemandResponse struct{}"
	}

	return strings.Join([]string{"UpdatePeriodToOnDemandResponse", string(data)}, " ")
}
