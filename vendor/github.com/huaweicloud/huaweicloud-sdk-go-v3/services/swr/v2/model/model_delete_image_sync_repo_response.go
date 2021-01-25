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
type DeleteImageSyncRepoResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteImageSyncRepoResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteImageSyncRepoResponse struct{}"
	}

	return strings.Join([]string{"DeleteImageSyncRepoResponse", string(data)}, " ")
}
