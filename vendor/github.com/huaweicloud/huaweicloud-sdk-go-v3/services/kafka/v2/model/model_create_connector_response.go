/*
 * Kafka
 *
 * Kafka Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type CreateConnectorResponse struct {
	// 任务ID。
	JobId *string `json:"job_id,omitempty"`
	// 实例转储ID。
	ConnectorId    *string `json:"connector_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CreateConnectorResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateConnectorResponse struct{}"
	}

	return strings.Join([]string{"CreateConnectorResponse", string(data)}, " ")
}
