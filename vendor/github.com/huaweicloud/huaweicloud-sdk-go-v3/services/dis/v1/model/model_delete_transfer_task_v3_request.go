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
type DeleteTransferTaskV3Request struct {
	StreamName string `json:"stream_name"`
	TaskName   string `json:"task_name"`
}

func (o DeleteTransferTaskV3Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteTransferTaskV3Request struct{}"
	}

	return strings.Join([]string{"DeleteTransferTaskV3Request", string(data)}, " ")
}
