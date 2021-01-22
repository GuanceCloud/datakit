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

type ClassroomCard struct {
	// 课堂ID
	ClassroomId string `json:"classroom_id"`
	// 课堂名称
	Name string `json:"name"`
	// 课堂描述
	Description string `json:"description"`
	// 课堂学分
	Credit float32 `json:"credit"`
	// 课堂当前的状态，normal：课堂处于正常状态，archive：课堂已归档
	Status string `json:"status"`
}

func (o ClassroomCard) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ClassroomCard struct{}"
	}

	return strings.Join([]string{"ClassroomCard", string(data)}, " ")
}
