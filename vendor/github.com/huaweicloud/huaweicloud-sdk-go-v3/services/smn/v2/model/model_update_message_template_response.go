/*
 * SMN
 *
 * SMN Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type UpdateMessageTemplateResponse struct {
	// 请求的唯一标识ID。
	RequestId      *string `json:"request_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o UpdateMessageTemplateResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateMessageTemplateResponse struct{}"
	}

	return strings.Join([]string{"UpdateMessageTemplateResponse", string(data)}, " ")
}
