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
type ShowCoordinatorsResponse struct {
	// 所有消费组对应的协调器列表。
	Coordinators   *[]ShowCoordinatorsRespCoordinators `json:"coordinators,omitempty"`
	HttpStatusCode int                                 `json:"-"`
}

func (o ShowCoordinatorsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowCoordinatorsResponse struct{}"
	}

	return strings.Join([]string{"ShowCoordinatorsResponse", string(data)}, " ")
}
