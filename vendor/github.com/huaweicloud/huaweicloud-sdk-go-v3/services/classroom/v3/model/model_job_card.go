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

type JobCard struct {
	// 作业名称
	Name string `json:"name"`
	// 作业ID
	JobId string `json:"job_id"`
	// 作业下发状态(unsend:作业未下发, send:作业已下发)
	IsSend string `json:"is_send"`
	// 作业截止时间, 日期格式：yyyy-MM-dd HH:mm:ss
	EndTime string `json:"end_time"`
	// 作业均分
	AverageScore string `json:"average_score"`
	// 作业提交人数
	SubmitJobNum int32 `json:"submit_job_num"`
	// 作业创建状态(yes:作业可以下发, no:作业不能下发)
	CreateStatus string `json:"create_status"`
	// 作业下发类型(specific:指定成员下发, all:下发课堂全员)
	SendType string `json:"send_type"`
	// 作业成绩是否公布(unpublish:表示未公布成绩, publish:表示已公布成绩)
	IsScoreVisibility string `json:"is_score_visibility"`
	// 作业下发时间, 日期格式：yyyy-MM-dd HH:mm:ss
	SendTime string `json:"send_time"`
}

func (o JobCard) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "JobCard struct{}"
	}

	return strings.Join([]string{"JobCard", string(data)}, " ")
}
