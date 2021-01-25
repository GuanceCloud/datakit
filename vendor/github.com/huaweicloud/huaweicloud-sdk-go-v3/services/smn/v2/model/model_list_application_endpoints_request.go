/*
 * SMN
 *
 * SMN Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ListApplicationEndpointsRequest struct {
	ApplicationUrn string  `json:"application_urn"`
	Offset         *int32  `json:"offset,omitempty"`
	Limit          *int32  `json:"limit,omitempty"`
	Enabled        *string `json:"enabled,omitempty"`
	Token          *string `json:"token,omitempty"`
	UserData       *string `json:"user_data,omitempty"`
}

func (o ListApplicationEndpointsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListApplicationEndpointsRequest struct{}"
	}

	return strings.Join([]string{"ListApplicationEndpointsRequest", string(data)}, " ")
}
