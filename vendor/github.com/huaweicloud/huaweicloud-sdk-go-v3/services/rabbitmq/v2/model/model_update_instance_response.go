/*
 * RabbitMQ
 *
 * RabbitMQ Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type UpdateInstanceResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o UpdateInstanceResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateInstanceResponse struct{}"
	}

	return strings.Join([]string{"UpdateInstanceResponse", string(data)}, " ")
}
