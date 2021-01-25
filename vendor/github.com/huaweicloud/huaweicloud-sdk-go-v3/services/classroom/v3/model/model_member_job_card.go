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

type MemberJobCard struct {
	// 作业名称
	Name string `json:"name"`
	// 作业均分(作业有均分该字段才返回)
	AverageScore string `json:"average_score"`
	// 作业得分(作业有分数该字段才返回)
	Score int32 `json:"score"`
	// 作业下发时间, 日期格式：yyyy-MM-dd HH:mm:ss
	SendTime string `json:"send_time"`
	// 作业最后一次提交时间, 日期格式：yyyy-MM-dd HH:mm:ss
	LastSubmitTime string `json:"last_submit_time"`
}

func (o MemberJobCard) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "MemberJobCard struct{}"
	}

	return strings.Join([]string{"MemberJobCard", string(data)}, " ")
}
