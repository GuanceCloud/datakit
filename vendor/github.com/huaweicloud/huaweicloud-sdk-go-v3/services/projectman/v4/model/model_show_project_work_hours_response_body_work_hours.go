/*
 * ProjectMan
 *
 * devcloud projectman api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ShowProjectWorkHoursResponseBodyWorkHours struct {
	// 项目名称
	ProjectName *string `json:"project_name,omitempty"`
	// 用户昵称
	NickName *string `json:"nick_name,omitempty"`
	// 用户名
	UserName *string `json:"user_name,omitempty"`
	// 工时日期
	WorkDate *string `json:"work_date,omitempty"`
	// 工时花费
	WorkHoursNum *string `json:"work_hours_num,omitempty"`
	// 工时内容
	Summary *string `json:"summary,omitempty"`
	// 工时类型
	WorkHoursTypeName *string `json:"work_hours_type_name,omitempty"`
	// 工作项编码
	IssueId *string `json:"issue_id,omitempty"`
	// 工作项类型
	IssueType *string `json:"issue_type,omitempty"`
	// 工作项标题
	Subject *string `json:"subject,omitempty"`
	// 工作项创建时间
	CreatedTime *string `json:"created_time,omitempty"`
	// 工作项结束时间
	ClosedTime *string `json:"closed_time,omitempty"`
}

func (o ShowProjectWorkHoursResponseBodyWorkHours) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowProjectWorkHoursResponseBodyWorkHours struct{}"
	}

	return strings.Join([]string{"ShowProjectWorkHoursResponseBodyWorkHours", string(data)}, " ")
}
