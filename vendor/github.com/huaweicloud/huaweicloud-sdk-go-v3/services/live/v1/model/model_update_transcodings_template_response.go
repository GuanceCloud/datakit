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

// Response Object
type UpdateTranscodingsTemplateResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o UpdateTranscodingsTemplateResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateTranscodingsTemplateResponse struct{}"
	}

	return strings.Join([]string{"UpdateTranscodingsTemplateResponse", string(data)}, " ")
}
