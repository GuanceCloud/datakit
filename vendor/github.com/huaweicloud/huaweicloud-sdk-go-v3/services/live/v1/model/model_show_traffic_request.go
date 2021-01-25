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
type ShowTrafficRequest struct {
	Domain    *string `json:"domain,omitempty"`
	StartTime *string `json:"start_time,omitempty"`
	EndTime   *string `json:"end_time,omitempty"`
	Step      *int32  `json:"step,omitempty"`
}

func (o ShowTrafficRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowTrafficRequest struct{}"
	}

	return strings.Join([]string{"ShowTrafficRequest", string(data)}, " ")
}
