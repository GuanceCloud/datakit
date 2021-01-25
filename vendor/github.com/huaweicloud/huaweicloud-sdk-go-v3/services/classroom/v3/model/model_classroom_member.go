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

type ClassroomMember struct {
	// 成员ID
	MemberId string `json:"member_id"`
	// 成员名称
	Name string `json:"name"`
	// 成员学号/工号
	Number string `json:"number"`
	// 成员所在班级的名字
	ClassName string `json:"class_name"`
	// 成员用户名
	UserName string `json:"user_name"`
	// 成员加入课堂时间，日期格式：yyyy-MM-dd HH:mm:ss
	JoinTime string `json:"join_time"`
	// 该成员已接收到的作业数量
	JobReceivedCount int32 `json:"job_received_count"`
	// 该成员已完成的作业数量
	JobFinishedCount int32 `json:"job_finished_count"`
	// 该成员作业完成率
	JobFinishedRate float32 `json:"job_finished_rate"`
}

func (o ClassroomMember) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ClassroomMember struct{}"
	}

	return strings.Join([]string{"ClassroomMember", string(data)}, " ")
}
