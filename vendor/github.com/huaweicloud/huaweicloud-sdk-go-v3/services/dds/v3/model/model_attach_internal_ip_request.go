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
type AttachInternalIpRequest struct {
	InstanceId string                       `json:"instance_id"`
	Body       *AttachInternalIpRequestBody `json:"body,omitempty"`
}

func (o AttachInternalIpRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AttachInternalIpRequest struct{}"
	}

	return strings.Join([]string{"AttachInternalIpRequest", string(data)}, " ")
}
