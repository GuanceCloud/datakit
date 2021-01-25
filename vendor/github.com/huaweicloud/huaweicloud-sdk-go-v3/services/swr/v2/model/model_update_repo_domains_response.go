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
type UpdateRepoDomainsResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o UpdateRepoDomainsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateRepoDomainsResponse struct{}"
	}

	return strings.Join([]string{"UpdateRepoDomainsResponse", string(data)}, " ")
}
