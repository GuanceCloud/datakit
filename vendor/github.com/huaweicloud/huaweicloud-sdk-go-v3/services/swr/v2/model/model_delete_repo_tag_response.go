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
type DeleteRepoTagResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteRepoTagResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteRepoTagResponse struct{}"
	}

	return strings.Join([]string{"DeleteRepoTagResponse", string(data)}, " ")
}
