/*
 * kps
 *
 * kps v3 版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ListRunningTaskRequest struct {
}

func (o ListRunningTaskRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListRunningTaskRequest struct{}"
	}

	return strings.Join([]string{"ListRunningTaskRequest", string(data)}, " ")
}
