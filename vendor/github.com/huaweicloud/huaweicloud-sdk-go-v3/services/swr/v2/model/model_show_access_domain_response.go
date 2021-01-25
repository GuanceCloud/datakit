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
type ShowAccessDomainResponse struct {
	// true：存在；false：不存在
	Exist          *bool `json:"exist,omitempty"`
	HttpStatusCode int   `json:"-"`
}

func (o ShowAccessDomainResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowAccessDomainResponse struct{}"
	}

	return strings.Join([]string{"ShowAccessDomainResponse", string(data)}, " ")
}
