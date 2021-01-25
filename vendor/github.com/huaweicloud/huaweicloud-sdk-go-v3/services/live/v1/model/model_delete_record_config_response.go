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
type DeleteRecordConfigResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteRecordConfigResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteRecordConfigResponse struct{}"
	}

	return strings.Join([]string{"DeleteRecordConfigResponse", string(data)}, " ")
}
