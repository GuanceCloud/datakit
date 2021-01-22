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
type UpdateBandwidthResponse struct {
	Bandwidth      *BandwidthResp `json:"bandwidth,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o UpdateBandwidthResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateBandwidthResponse struct{}"
	}

	return strings.Join([]string{"UpdateBandwidthResponse", string(data)}, " ")
}
