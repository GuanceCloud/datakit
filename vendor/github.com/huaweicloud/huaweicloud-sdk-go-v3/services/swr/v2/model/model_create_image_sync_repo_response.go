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
type CreateImageSyncRepoResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CreateImageSyncRepoResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateImageSyncRepoResponse struct{}"
	}

	return strings.Join([]string{"CreateImageSyncRepoResponse", string(data)}, " ")
}
