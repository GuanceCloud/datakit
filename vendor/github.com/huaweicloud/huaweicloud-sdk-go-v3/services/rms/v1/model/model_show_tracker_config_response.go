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

// Response Object
type ShowTrackerConfigResponse struct {
	Channel  *ChannelConfigBody  `json:"channel,omitempty"`
	Selector *SelectorConfigBody `json:"selector,omitempty"`
	// IAM委托名称
	AgencyName     *string `json:"agency_name,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ShowTrackerConfigResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowTrackerConfigResponse struct{}"
	}

	return strings.Join([]string{"ShowTrackerConfigResponse", string(data)}, " ")
}
