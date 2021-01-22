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

// Request Object
type DeleteDependencyRequest struct {
	DependId string `json:"depend_id"`
}

func (o DeleteDependencyRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteDependencyRequest struct{}"
	}

	return strings.Join([]string{"DeleteDependencyRequest", string(data)}, " ")
}
