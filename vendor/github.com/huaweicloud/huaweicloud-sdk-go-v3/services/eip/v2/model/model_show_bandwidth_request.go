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

// Request Object
type ShowBandwidthRequest struct {
	BandwidthId string `json:"bandwidth_id"`
}

func (o ShowBandwidthRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowBandwidthRequest struct{}"
	}

	return strings.Join([]string{"ShowBandwidthRequest", string(data)}, " ")
}
