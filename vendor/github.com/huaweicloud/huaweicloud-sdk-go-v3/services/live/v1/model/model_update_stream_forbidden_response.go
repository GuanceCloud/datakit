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
type UpdateStreamForbiddenResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o UpdateStreamForbiddenResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateStreamForbiddenResponse struct{}"
	}

	return strings.Join([]string{"UpdateStreamForbiddenResponse", string(data)}, " ")
}
