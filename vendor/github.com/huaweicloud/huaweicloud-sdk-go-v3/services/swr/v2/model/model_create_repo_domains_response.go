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
type CreateRepoDomainsResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CreateRepoDomainsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateRepoDomainsResponse struct{}"
	}

	return strings.Join([]string{"CreateRepoDomainsResponse", string(data)}, " ")
}
