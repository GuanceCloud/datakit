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
type StartFailoverResponse struct {
	// 实例Id
	InstanceId *string `json:"instance_id,omitempty"`
	// 节点Id
	NodeId *string `json:"nodeId,omitempty"`
	// 任务Id
	WorkflowId     *string `json:"workflowId,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o StartFailoverResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "StartFailoverResponse struct{}"
	}

	return strings.Join([]string{"StartFailoverResponse", string(data)}, " ")
}
