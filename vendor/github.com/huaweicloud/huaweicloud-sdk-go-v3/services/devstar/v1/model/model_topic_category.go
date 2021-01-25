/*
 * DevStar
 *
 * DevStar API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type TopicCategory struct {
	// topic的id
	TopicId *string `json:"topic_id,omitempty"`
	// topic的名称
	TopicName *string `json:"topic_name,omitempty"`
	// topic对应的类别的id
	CategoryId *string `json:"category_id,omitempty"`
	// topic对应的类别的名称
	CategoryName *string `json:"category_name,omitempty"`
}

func (o TopicCategory) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TopicCategory struct{}"
	}

	return strings.Join([]string{"TopicCategory", string(data)}, " ")
}
