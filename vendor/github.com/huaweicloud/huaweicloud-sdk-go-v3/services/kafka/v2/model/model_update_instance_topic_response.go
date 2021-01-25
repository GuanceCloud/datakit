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
type UpdateInstanceTopicResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o UpdateInstanceTopicResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateInstanceTopicResponse struct{}"
	}

	return strings.Join([]string{"UpdateInstanceTopicResponse", string(data)}, " ")
}
