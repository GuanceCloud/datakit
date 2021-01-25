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
type AttachEipResponse struct {
	// 任务ID。
	JobId *string `json:"job_id,omitempty"`
	// 节点ID。
	NodeId *string `json:"node_id,omitempty"`
	// 节点名称。
	NodeName *string `json:"node_name,omitempty"`
	// 公网IP的ID。
	PublicIpId *string `json:"public_ip_id,omitempty"`
	// 公网IP。
	PublicIp       *string `json:"public_ip,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o AttachEipResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AttachEipResponse struct{}"
	}

	return strings.Join([]string{"AttachEipResponse", string(data)}, " ")
}
