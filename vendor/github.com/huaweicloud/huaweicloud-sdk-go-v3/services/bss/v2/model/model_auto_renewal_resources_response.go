/*
 * BSS
 *
 * Business Support System API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type AutoRenewalResourcesResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o AutoRenewalResourcesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AutoRenewalResourcesResponse struct{}"
	}

	return strings.Join([]string{"AutoRenewalResourcesResponse", string(data)}, " ")
}
