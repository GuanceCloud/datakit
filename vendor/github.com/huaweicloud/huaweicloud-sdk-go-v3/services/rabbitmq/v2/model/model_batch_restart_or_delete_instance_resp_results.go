/*
 * RabbitMQ
 *
 * RabbitMQ Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type BatchRestartOrDeleteInstanceRespResults struct {
	// 操作结果：   - success: 操作成功   - failed: 操作失败
	Result *string `json:"result,omitempty"`
	// 实例ID。
	Instance *string `json:"instance,omitempty"`
}

func (o BatchRestartOrDeleteInstanceRespResults) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchRestartOrDeleteInstanceRespResults struct{}"
	}

	return strings.Join([]string{"BatchRestartOrDeleteInstanceRespResults", string(data)}, " ")
}
