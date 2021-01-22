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
type DeleteRepoDomainsResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteRepoDomainsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteRepoDomainsResponse struct{}"
	}

	return strings.Join([]string{"DeleteRepoDomainsResponse", string(data)}, " ")
}
