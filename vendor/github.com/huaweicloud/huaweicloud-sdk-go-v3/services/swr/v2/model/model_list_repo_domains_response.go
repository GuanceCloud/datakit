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
type ListRepoDomainsResponse struct {
	// 共享租户列表
	Body           *[]ShowRepoDomainsResponse `json:"body,omitempty"`
	HttpStatusCode int                        `json:"-"`
}

func (o ListRepoDomainsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListRepoDomainsResponse struct{}"
	}

	return strings.Join([]string{"ListRepoDomainsResponse", string(data)}, " ")
}
