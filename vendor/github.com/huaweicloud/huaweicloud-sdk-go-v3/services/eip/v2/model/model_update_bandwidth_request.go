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
type UpdateBandwidthRequest struct {
	BandwidthId string                      `json:"bandwidth_id"`
	Body        *UpdateBandwidthRequestBody `json:"body,omitempty"`
}

func (o UpdateBandwidthRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateBandwidthRequest struct{}"
	}

	return strings.Join([]string{"UpdateBandwidthRequest", string(data)}, " ")
}
