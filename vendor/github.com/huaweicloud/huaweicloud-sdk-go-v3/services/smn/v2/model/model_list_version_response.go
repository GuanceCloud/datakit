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

// Response Object
type ListVersionResponse struct {
	Version        *interface{} `json:"version,omitempty"`
	HttpStatusCode int          `json:"-"`
}

func (o ListVersionResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListVersionResponse struct{}"
	}

	return strings.Join([]string{"ListVersionResponse", string(data)}, " ")
}
