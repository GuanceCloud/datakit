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
type DeleteTranscodingsTemplateRequest struct {
	Domain  string `json:"domain"`
	AppName string `json:"app_name"`
}

func (o DeleteTranscodingsTemplateRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteTranscodingsTemplateRequest struct{}"
	}

	return strings.Join([]string{"DeleteTranscodingsTemplateRequest", string(data)}, " ")
}
