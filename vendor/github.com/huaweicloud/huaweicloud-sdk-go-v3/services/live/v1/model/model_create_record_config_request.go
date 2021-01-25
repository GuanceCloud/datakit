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
type CreateRecordConfigRequest struct {
	Body *RecordConfigInfo `json:"body,omitempty"`
}

func (o CreateRecordConfigRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateRecordConfigRequest struct{}"
	}

	return strings.Join([]string{"CreateRecordConfigRequest", string(data)}, " ")
}
