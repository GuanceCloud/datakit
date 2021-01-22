/*
 * SWR
 *
 * SWR API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListRetentionsResponse struct {
	Body           *[]Retention `json:"body,omitempty"`
	HttpStatusCode int          `json:"-"`
}

func (o ListRetentionsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListRetentionsResponse struct{}"
	}

	return strings.Join([]string{"ListRetentionsResponse", string(data)}, " ")
}
