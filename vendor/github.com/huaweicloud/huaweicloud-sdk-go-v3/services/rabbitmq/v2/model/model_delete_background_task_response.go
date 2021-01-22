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
type DeleteBackgroundTaskResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteBackgroundTaskResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteBackgroundTaskResponse struct{}"
	}

	return strings.Join([]string{"DeleteBackgroundTaskResponse", string(data)}, " ")
}
