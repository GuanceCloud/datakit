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
type UpdateInstanceAutoCreateTopicResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o UpdateInstanceAutoCreateTopicResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateInstanceAutoCreateTopicResponse struct{}"
	}

	return strings.Join([]string{"UpdateInstanceAutoCreateTopicResponse", string(data)}, " ")
}
