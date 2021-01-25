/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ChangeFailoverModeResponse struct {
	// 实例Id
	InstanceId *string `json:"instanceId,omitempty"`
	// 同步模式
	ReplicationMode *string `json:"replicationMode,omitempty"`
	// 任务id
	WorkflowId     *string `json:"workflowId,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ChangeFailoverModeResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ChangeFailoverModeResponse struct{}"
	}

	return strings.Join([]string{"ChangeFailoverModeResponse", string(data)}, " ")
}
