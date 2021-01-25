/*
 * AS
 *
 * 弹性伸缩API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type CompleteLifecycleActionResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CompleteLifecycleActionResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CompleteLifecycleActionResponse struct{}"
	}

	return strings.Join([]string{"CompleteLifecycleActionResponse", string(data)}, " ")
}
