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
type ShowGroupsResponse struct {
	Group          *ShowGroupsRespGroup `json:"group,omitempty"`
	HttpStatusCode int                  `json:"-"`
}

func (o ShowGroupsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowGroupsResponse struct{}"
	}

	return strings.Join([]string{"ShowGroupsResponse", string(data)}, " ")
}
