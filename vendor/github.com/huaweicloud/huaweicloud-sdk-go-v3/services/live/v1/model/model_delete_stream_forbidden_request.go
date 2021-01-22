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
type DeleteStreamForbiddenRequest struct {
	SpecifyProject *string `json:"specify_project,omitempty"`
	Domain         string  `json:"domain"`
	AppName        string  `json:"app_name"`
	StreamName     string  `json:"stream_name"`
}

func (o DeleteStreamForbiddenRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteStreamForbiddenRequest struct{}"
	}

	return strings.Join([]string{"DeleteStreamForbiddenRequest", string(data)}, " ")
}
