/*
 * Classroom
 *
 * devcloud classedge api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type JobRecords struct {
	// 第XX次提交
	Name string `json:"name"`
	// 习题判题得分
	AutoScore int32 `json:"auto_score"`
	// 习题用例通过数
	CasePassCount int32 `json:"case_pass_count"`
	// 习题用例总数
	ExeCaseCount int32 `json:"exe_case_count"`
	// 代码行数
	CodeLine int32 `json:"code_line"`
	// 习题提交时间, 日期格式：yyyy-MM-dd HH:mm:ss
	CommitTime string `json:"commit_time"`
	// 习题圈复杂度
	ComplexityFileAvg string `json:"complexity_file_avg"`
	// 习题判题耗时(毫秒)
	AutoScoreUsingTime int32 `json:"auto_score_using_time"`
}

func (o JobRecords) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "JobRecords struct{}"
	}

	return strings.Join([]string{"JobRecords", string(data)}, " ")
}
