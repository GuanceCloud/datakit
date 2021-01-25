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
type DeleteSharedBandwidthRequest struct {
	BandwidthId string `json:"bandwidth_id"`
}

func (o DeleteSharedBandwidthRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteSharedBandwidthRequest struct{}"
	}

	return strings.Join([]string{"DeleteSharedBandwidthRequest", string(data)}, " ")
}
