/*
 * EIP
 *
 * 云服务接口
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type CreateSharedBandwidthResponse struct {
	Bandwidth      *BandwidthResp `json:"bandwidth,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o CreateSharedBandwidthResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateSharedBandwidthResponse struct{}"
	}

	return strings.Join([]string{"CreateSharedBandwidthResponse", string(data)}, " ")
}
