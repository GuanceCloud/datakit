/*
 * DDS
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
type CancelEipResponse struct {
	// 任务ID。
	JobId *string `json:"job_id,omitempty"`
	// 节点ID。
	NodeId *string `json:"node_id,omitempty"`
	// 节点名称。
	NodeName       *string `json:"node_name,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CancelEipResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CancelEipResponse struct{}"
	}

	return strings.Join([]string{"CancelEipResponse", string(data)}, " ")
}
