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
type CreateTranscodingsTemplateRequest struct {
	Body *StreamTranscodingTemplate `json:"body,omitempty"`
}

func (o CreateTranscodingsTemplateRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateTranscodingsTemplateRequest struct{}"
	}

	return strings.Join([]string{"CreateTranscodingsTemplateRequest", string(data)}, " ")
}
