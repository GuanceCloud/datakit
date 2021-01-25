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
type BatchCreateOrDeleteInstanceTagResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o BatchCreateOrDeleteInstanceTagResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchCreateOrDeleteInstanceTagResponse struct{}"
	}

	return strings.Join([]string{"BatchCreateOrDeleteInstanceTagResponse", string(data)}, " ")
}
