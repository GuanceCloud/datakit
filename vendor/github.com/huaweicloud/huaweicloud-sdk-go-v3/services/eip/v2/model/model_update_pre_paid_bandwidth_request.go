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
type UpdatePrePaidBandwidthRequest struct {
	BandwidthId string                             `json:"bandwidth_id"`
	Body        *UpdatePrePaidBandwidthRequestBody `json:"body,omitempty"`
}

func (o UpdatePrePaidBandwidthRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdatePrePaidBandwidthRequest struct{}"
	}

	return strings.Join([]string{"UpdatePrePaidBandwidthRequest", string(data)}, " ")
}
