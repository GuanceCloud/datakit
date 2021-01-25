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

type ExerciseGroup struct {
	// 习题列表
	Exercises []ExerciseCard `json:"exercises"`
	// 习题分类
	Type string `json:"type"`
}

func (o ExerciseGroup) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ExerciseGroup struct{}"
	}

	return strings.Join([]string{"ExerciseGroup", string(data)}, " ")
}
