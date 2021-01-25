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

// Response Object
type ListTransferTasksV3Response struct {
	// 转储任务总数。
	TotalNumber *int32 `json:"total_number,omitempty"`
	// 转储任务列表。
	Tasks          *[]TransferTask `json:"tasks,omitempty"`
	HttpStatusCode int             `json:"-"`
}

func (o ListTransferTasksV3Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListTransferTasksV3Response struct{}"
	}

	return strings.Join([]string{"ListTransferTasksV3Response", string(data)}, " ")
}
