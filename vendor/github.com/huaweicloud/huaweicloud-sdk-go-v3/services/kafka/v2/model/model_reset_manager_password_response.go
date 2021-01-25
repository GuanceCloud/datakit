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
type ResetManagerPasswordResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o ResetManagerPasswordResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResetManagerPasswordResponse struct{}"
	}

	return strings.Join([]string{"ResetManagerPasswordResponse", string(data)}, " ")
}
