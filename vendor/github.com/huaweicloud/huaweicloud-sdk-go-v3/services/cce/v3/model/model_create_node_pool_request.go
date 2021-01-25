/*
 * CCE
 *
 * CCE开放API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type CreateNodePoolRequest struct {
	ContentType string    `json:"Content-Type"`
	ClusterId   string    `json:"cluster_id"`
	Body        *NodePool `json:"body,omitempty"`
}

func (o CreateNodePoolRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateNodePoolRequest struct{}"
	}

	return strings.Join([]string{"CreateNodePoolRequest", string(data)}, " ")
}
