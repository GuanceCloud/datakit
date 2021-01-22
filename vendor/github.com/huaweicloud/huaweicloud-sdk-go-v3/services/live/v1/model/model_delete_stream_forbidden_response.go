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
type DeleteStreamForbiddenResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteStreamForbiddenResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteStreamForbiddenResponse struct{}"
	}

	return strings.Join([]string{"DeleteStreamForbiddenResponse", string(data)}, " ")
}
