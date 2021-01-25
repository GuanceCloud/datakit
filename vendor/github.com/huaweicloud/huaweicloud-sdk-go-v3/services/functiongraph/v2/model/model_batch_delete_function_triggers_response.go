/*
 * FunctionGraph
 *
 * API v2
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type BatchDeleteFunctionTriggersResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o BatchDeleteFunctionTriggersResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchDeleteFunctionTriggersResponse struct{}"
	}

	return strings.Join([]string{"BatchDeleteFunctionTriggersResponse", string(data)}, " ")
}
