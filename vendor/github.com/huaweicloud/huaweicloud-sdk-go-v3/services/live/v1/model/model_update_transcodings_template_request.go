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
type UpdateTranscodingsTemplateRequest struct {
	Body *StreamTranscodingTemplate `json:"body,omitempty"`
}

func (o UpdateTranscodingsTemplateRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateTranscodingsTemplateRequest struct{}"
	}

	return strings.Join([]string{"UpdateTranscodingsTemplateRequest", string(data)}, " ")
}
