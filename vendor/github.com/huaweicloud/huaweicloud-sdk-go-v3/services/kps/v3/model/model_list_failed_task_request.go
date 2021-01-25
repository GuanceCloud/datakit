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
type ListFailedTaskRequest struct {
}

func (o ListFailedTaskRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListFailedTaskRequest struct{}"
	}

	return strings.Join([]string{"ListFailedTaskRequest", string(data)}, " ")
}
