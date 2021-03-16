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
type ShowResourceHistoryResponse struct {
	// 资源历史列表
	Items          *[]HistoryItem `json:"items,omitempty"`
	PageInfo       *PageInfo      `json:"page_info,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o ShowResourceHistoryResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowResourceHistoryResponse struct{}"
	}

	return strings.Join([]string{"ShowResourceHistoryResponse", string(data)}, " ")
}
