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
type ListServerInterfacesRequest struct {
	ServerId string `json:"server_id"`
}

func (o ListServerInterfacesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListServerInterfacesRequest struct{}"
	}

	return strings.Join([]string{"ListServerInterfacesRequest", string(data)}, " ")
}
