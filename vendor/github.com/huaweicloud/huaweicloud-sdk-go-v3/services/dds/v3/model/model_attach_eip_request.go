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
type AttachEipRequest struct {
	NodeId string                `json:"node_id"`
	Body   *AttachEipRequestBody `json:"body,omitempty"`
}

func (o AttachEipRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AttachEipRequest struct{}"
	}

	return strings.Join([]string{"AttachEipRequest", string(data)}, " ")
}
