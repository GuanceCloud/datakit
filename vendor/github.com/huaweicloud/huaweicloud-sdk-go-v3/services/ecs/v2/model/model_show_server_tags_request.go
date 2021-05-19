/*
 * ECS
 *
 * ECS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ShowServerTagsRequest struct {
	ServerId string `json:"server_id"`
}

func (o ShowServerTagsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowServerTagsRequest struct{}"
	}

	return strings.Join([]string{"ShowServerTagsRequest", string(data)}, " ")
}
