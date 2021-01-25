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
type ShowServerLimitsRequest struct {
}

func (o ShowServerLimitsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowServerLimitsRequest struct{}"
	}

	return strings.Join([]string{"ShowServerLimitsRequest", string(data)}, " ")
}
