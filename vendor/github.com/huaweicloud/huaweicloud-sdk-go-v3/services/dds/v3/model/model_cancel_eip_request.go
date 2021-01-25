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

// Request Object
type CancelEipRequest struct {
	NodeId string `json:"node_id"`
}

func (o CancelEipRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CancelEipRequest struct{}"
	}

	return strings.Join([]string{"CancelEipRequest", string(data)}, " ")
}
