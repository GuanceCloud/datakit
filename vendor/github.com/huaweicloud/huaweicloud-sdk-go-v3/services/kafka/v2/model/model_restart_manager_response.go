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
type RestartManagerResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o RestartManagerResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RestartManagerResponse struct{}"
	}

	return strings.Join([]string{"RestartManagerResponse", string(data)}, " ")
}
