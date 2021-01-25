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
type UpdateVersionAliasRequest struct {
	FunctionUrn string                         `json:"function_urn"`
	Name        string                         `json:"name"`
	Body        *UpdateVersionAliasRequestBody `json:"body,omitempty"`
}

func (o UpdateVersionAliasRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateVersionAliasRequest struct{}"
	}

	return strings.Join([]string{"UpdateVersionAliasRequest", string(data)}, " ")
}
