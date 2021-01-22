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
type CreateStreamForbiddenResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CreateStreamForbiddenResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateStreamForbiddenResponse struct{}"
	}

	return strings.Join([]string{"CreateStreamForbiddenResponse", string(data)}, " ")
}
