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
type BatchStartBaremetalServersResponse struct {
	// 提交任务成功后返回的任务ID，用户可以使用该ID对任务执行情况进行查询
	JobId          *string `json:"job_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o BatchStartBaremetalServersResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchStartBaremetalServersResponse struct{}"
	}

	return strings.Join([]string{"BatchStartBaremetalServersResponse", string(data)}, " ")
}
