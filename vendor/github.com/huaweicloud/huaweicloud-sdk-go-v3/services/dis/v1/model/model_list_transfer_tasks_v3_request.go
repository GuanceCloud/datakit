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
type ListTransferTasksV3Request struct {
	StreamName string `json:"stream_name"`
}

func (o ListTransferTasksV3Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListTransferTasksV3Request struct{}"
	}

	return strings.Join([]string{"ListTransferTasksV3Request", string(data)}, " ")
}
