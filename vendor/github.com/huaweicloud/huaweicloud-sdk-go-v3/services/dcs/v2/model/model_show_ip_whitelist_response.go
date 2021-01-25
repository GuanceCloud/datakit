/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ShowIpWhitelistResponse struct {
	// 是否启用白名单（true/false）。
	EnableWhitelist *bool `json:"enable_whitelist,omitempty"`
	// IP白名单分组列表。
	Whitelist      *[]Whitelist `json:"whitelist,omitempty"`
	HttpStatusCode int          `json:"-"`
}

func (o ShowIpWhitelistResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowIpWhitelistResponse struct{}"
	}

	return strings.Join([]string{"ShowIpWhitelistResponse", string(data)}, " ")
}
