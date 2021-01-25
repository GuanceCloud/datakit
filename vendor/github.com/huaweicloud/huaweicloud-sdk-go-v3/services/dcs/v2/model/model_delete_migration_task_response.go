/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type DeleteMigrationTaskResponse struct {
	// 删除的迁移任务ID列表。
	TaskIdList     *[]string `json:"task_id_list,omitempty"`
	HttpStatusCode int       `json:"-"`
}

func (o DeleteMigrationTaskResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteMigrationTaskResponse struct{}"
	}

	return strings.Join([]string{"DeleteMigrationTaskResponse", string(data)}, " ")
}
