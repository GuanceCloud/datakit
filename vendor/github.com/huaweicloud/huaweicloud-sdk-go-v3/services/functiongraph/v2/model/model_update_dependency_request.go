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
type UpdateDependencyRequest struct {
	DependId string                       `json:"depend_id"`
	Body     *UpdateDependencyRequestBody `json:"body,omitempty"`
}

func (o UpdateDependencyRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateDependencyRequest struct{}"
	}

	return strings.Join([]string{"UpdateDependencyRequest", string(data)}, " ")
}
