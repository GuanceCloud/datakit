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
type AddPublicipsIntoSharedBandwidthRequest struct {
	BandwidthId string                                      `json:"bandwidth_id"`
	Body        *AddPublicipsIntoSharedBandwidthRequestBody `json:"body,omitempty"`
}

func (o AddPublicipsIntoSharedBandwidthRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AddPublicipsIntoSharedBandwidthRequest struct{}"
	}

	return strings.Join([]string{"AddPublicipsIntoSharedBandwidthRequest", string(data)}, " ")
}
