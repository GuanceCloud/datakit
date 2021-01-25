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
type DownloadSlowlogRequest struct {
	InstanceId string                      `json:"instance_id"`
	Body       *DownloadSlowlogRequestBody `json:"body,omitempty"`
}

func (o DownloadSlowlogRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DownloadSlowlogRequest struct{}"
	}

	return strings.Join([]string{"DownloadSlowlogRequest", string(data)}, " ")
}
