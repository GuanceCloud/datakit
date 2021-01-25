/*
 * Kafka
 *
 * Kafka Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type CreateInstanceTopicResponse struct {
	// topic名称。
	Name           *string `json:"name,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CreateInstanceTopicResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateInstanceTopicResponse struct{}"
	}

	return strings.Join([]string{"CreateInstanceTopicResponse", string(data)}, " ")
}
