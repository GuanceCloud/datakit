/*
 * RMS
 *
 * Resource Manager Api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// tracker通道配置
type ChannelConfigBody struct {
	Smn *TrackerSmnChannelConfigBody `json:"smn,omitempty"`
	Obs *TrackerObsChannelConfigBody `json:"obs,omitempty"`
}

func (o ChannelConfigBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ChannelConfigBody struct{}"
	}

	return strings.Join([]string{"ChannelConfigBody", string(data)}, " ")
}
