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
type CreatePartitionResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CreatePartitionResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreatePartitionResponse struct{}"
	}

	return strings.Join([]string{"CreatePartitionResponse", string(data)}, " ")
}
