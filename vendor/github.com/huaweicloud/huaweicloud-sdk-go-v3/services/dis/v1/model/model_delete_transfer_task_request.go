/*
 * DIS
 *
 * DIS v1 API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type DeleteTransferTaskRequest struct {
	StreamName string `json:"stream_name"`
	TaskName   string `json:"task_name"`
}

func (o DeleteTransferTaskRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteTransferTaskRequest struct{}"
	}

	return strings.Join([]string{"DeleteTransferTaskRequest", string(data)}, " ")
}
