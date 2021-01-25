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
type ShowBandwidthResponse struct {
	Bandwidth      *BandwidthResp `json:"bandwidth,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o ShowBandwidthResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowBandwidthResponse struct{}"
	}

	return strings.Join([]string{"ShowBandwidthResponse", string(data)}, " ")
}
