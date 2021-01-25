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
type DeleteRepoResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteRepoResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteRepoResponse struct{}"
	}

	return strings.Join([]string{"DeleteRepoResponse", string(data)}, " ")
}
