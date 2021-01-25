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
type ListApplicationAttributesRequest struct {
	ApplicationUrn string `json:"application_urn"`
}

func (o ListApplicationAttributesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListApplicationAttributesRequest struct{}"
	}

	return strings.Join([]string{"ListApplicationAttributesRequest", string(data)}, " ")
}
