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
type DeleteRecordConfigRequest struct {
	Domain  string `json:"domain"`
	AppName string `json:"app_name"`
}

func (o DeleteRecordConfigRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteRecordConfigRequest struct{}"
	}

	return strings.Join([]string{"DeleteRecordConfigRequest", string(data)}, " ")
}
