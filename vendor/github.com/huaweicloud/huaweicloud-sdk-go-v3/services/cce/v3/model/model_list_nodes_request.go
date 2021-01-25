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
type ListNodesRequest struct {
	ClusterId   string  `json:"cluster_id"`
	ContentType string  `json:"Content-Type"`
	ErrorStatus *string `json:"errorStatus,omitempty"`
}

func (o ListNodesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListNodesRequest struct{}"
	}

	return strings.Join([]string{"ListNodesRequest", string(data)}, " ")
}
