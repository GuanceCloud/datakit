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
type CreateTranscodingsTemplateResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CreateTranscodingsTemplateResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateTranscodingsTemplateResponse struct{}"
	}

	return strings.Join([]string{"CreateTranscodingsTemplateResponse", string(data)}, " ")
}
