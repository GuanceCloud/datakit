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
type DeleteAllFailedTaskRequest struct {
}

func (o DeleteAllFailedTaskRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteAllFailedTaskRequest struct{}"
	}

	return strings.Join([]string{"DeleteAllFailedTaskRequest", string(data)}, " ")
}
