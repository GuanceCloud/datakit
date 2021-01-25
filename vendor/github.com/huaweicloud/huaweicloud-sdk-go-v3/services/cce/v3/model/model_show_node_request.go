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
type ShowNodeRequest struct {
	ClusterId   string  `json:"cluster_id"`
	NodeId      string  `json:"node_id"`
	ContentType string  `json:"Content-Type"`
	ErrorStatus *string `json:"errorStatus,omitempty"`
}

func (o ShowNodeRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowNodeRequest struct{}"
	}

	return strings.Join([]string{"ShowNodeRequest", string(data)}, " ")
}
