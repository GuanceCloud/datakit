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
type ListStreamForbiddenRequest struct {
	SpecifyProject *string `json:"specify_project,omitempty"`
	Domain         string  `json:"domain"`
	AppName        *string `json:"app_name,omitempty"`
	StreamName     *string `json:"stream_name,omitempty"`
	Page           *int32  `json:"page,omitempty"`
	Size           *int32  `json:"size,omitempty"`
}

func (o ListStreamForbiddenRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListStreamForbiddenRequest struct{}"
	}

	return strings.Join([]string{"ListStreamForbiddenRequest", string(data)}, " ")
}
