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
type DeleteTranscodingsTemplateResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteTranscodingsTemplateResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteTranscodingsTemplateResponse struct{}"
	}

	return strings.Join([]string{"DeleteTranscodingsTemplateResponse", string(data)}, " ")
}
