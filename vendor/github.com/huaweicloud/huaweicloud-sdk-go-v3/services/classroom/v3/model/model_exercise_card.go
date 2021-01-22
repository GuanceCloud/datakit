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

type ExerciseCard struct {
	// 习题名称
	Name string `json:"name"`
	// 习题ID
	ExerciseId string `json:"exercise_id"`
	// 习题描述
	Description string `json:"description"`
	// 习题子类型 1：函数c 2：函数c++ 3：函数Java 4：函数Python 5：单人项目java 6：单人项目Hadoop 7：通用 8：企业级软件项目 10：web习题 11：AI习题 12：单选题 13：多选题 14：填空题 15：单人项目C 16：单人项目C++
	ResourceSubType int32 `json:"resource_sub_type"`
	// 习题分值
	TargetScore int32 `json:"target_score"`
}

func (o ExerciseCard) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ExerciseCard struct{}"
	}

	return strings.Join([]string{"ExerciseCard", string(data)}, " ")
}
