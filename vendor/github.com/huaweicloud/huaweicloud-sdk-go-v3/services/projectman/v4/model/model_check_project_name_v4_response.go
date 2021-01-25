/*
 * ProjectMan
 *
 * devcloud projectman api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type CheckProjectNameV4Response struct {
	// 是否存在相同的项目名称 true 存在， false 不存在
	Exist          *bool `json:"exist,omitempty"`
	HttpStatusCode int   `json:"-"`
}

func (o CheckProjectNameV4Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CheckProjectNameV4Response struct{}"
	}

	return strings.Join([]string{"CheckProjectNameV4Response", string(data)}, " ")
}
