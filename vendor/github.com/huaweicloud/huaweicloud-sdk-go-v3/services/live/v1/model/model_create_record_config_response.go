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
type CreateRecordConfigResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CreateRecordConfigResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateRecordConfigResponse struct{}"
	}

	return strings.Join([]string{"CreateRecordConfigResponse", string(data)}, " ")
}
