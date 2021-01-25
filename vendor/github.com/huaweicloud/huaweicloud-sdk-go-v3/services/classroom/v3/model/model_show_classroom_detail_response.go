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

// Response Object
type ShowClassroomDetailResponse struct {
	// 课堂名称
	Name *string `json:"name,omitempty"`
	// 课堂描述
	Description *string `json:"description,omitempty"`
	// 课堂公告
	Announcement *string `json:"announcement,omitempty"`
	// 课堂公告创建时间，日期格式：yyyy-MM-dd
	AnnouncementTime *string `json:"announcement_time,omitempty"`
	// 课堂创建时间，日期格式：yyyy-MM-dd HH:mm:ss
	CreateTime *string `json:"create_time,omitempty"`
	// 课堂最新更新时间，日期格式：yyyy-MM-dd HH:mm:ss
	UpdateTime *string `json:"update_time,omitempty"`
	// 当前课堂的授课人
	Teacher *string `json:"teacher,omitempty"`
	// 课堂学分
	Credit float32 `json:"credit,omitempty"`
	// 课堂开始时间，日期格式：yyyy-MM-dd HH:mm:ss
	StartTime *string `json:"start_time,omitempty"`
	// 课堂结束时间，日期格式：yyyy-MM-dd HH:mm:ss
	EndTime *string `json:"end_time,omitempty"`
	// 当前用户在课堂下角色，取值范围：teacher：老师，student：学生
	Role *string `json:"role,omitempty"`
	// 授课学校
	School *string `json:"school,omitempty"`
	// 课堂下目录数量
	ContentCount *int32 `json:"content_count,omitempty"`
	// 课堂下课件数量
	CoursewareCount *int32 `json:"courseware_count,omitempty"`
	// 课堂下作业数量
	JobCount *int32 `json:"job_count,omitempty"`
	// 课堂下成员数量
	MemberCount *int32 `json:"member_count,omitempty"`
	// 课堂当前的状态，normal：课堂处于正常状态，archive：课堂已归档
	Status         *string `json:"status,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ShowClassroomDetailResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowClassroomDetailResponse struct{}"
	}

	return strings.Join([]string{"ShowClassroomDetailResponse", string(data)}, " ")
}
