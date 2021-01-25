/*
 * SWR
 *
 * SWR API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type UpdateTriggerRequestBody struct {
	// 是否生效,true启用，false不启用
	Enable string `json:"enable"`
}

func (o UpdateTriggerRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateTriggerRequestBody struct{}"
	}

	return strings.Join([]string{"UpdateTriggerRequestBody", string(data)}, " ")
}
