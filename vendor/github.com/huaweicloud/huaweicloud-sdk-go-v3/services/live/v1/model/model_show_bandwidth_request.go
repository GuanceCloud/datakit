/*
 * Live
 *
 * 直播服务源站所有接口
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ShowBandwidthRequest struct {
	Domain    *string `json:"domain,omitempty"`
	StartTime *string `json:"start_time,omitempty"`
	EndTime   *string `json:"end_time,omitempty"`
	Step      *int32  `json:"step,omitempty"`
}

func (o ShowBandwidthRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowBandwidthRequest struct{}"
	}

	return strings.Join([]string{"ShowBandwidthRequest", string(data)}, " ")
}
