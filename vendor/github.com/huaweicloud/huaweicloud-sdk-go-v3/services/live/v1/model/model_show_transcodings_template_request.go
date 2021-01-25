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
type ShowTranscodingsTemplateRequest struct {
	Domain  string  `json:"domain"`
	AppName *string `json:"app_name,omitempty"`
	Page    *int32  `json:"page,omitempty"`
	Size    *int32  `json:"size,omitempty"`
}

func (o ShowTranscodingsTemplateRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowTranscodingsTemplateRequest struct{}"
	}

	return strings.Join([]string{"ShowTranscodingsTemplateRequest", string(data)}, " ")
}
