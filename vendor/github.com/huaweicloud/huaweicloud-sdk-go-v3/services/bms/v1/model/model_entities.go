/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// entities字段数据结构说明
type Entities struct {
	// 子任务数量。没有子任务时为0
	SubJobsTotal *int32 `json:"sub_jobs_total,omitempty"`
	// 每个子任务的执行信息。没有子任务时为空列表
	SubJobs *[]SubJobs `json:"sub_jobs,omitempty"`
}

func (o Entities) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Entities struct{}"
	}

	return strings.Join([]string{"Entities", string(data)}, " ")
}
