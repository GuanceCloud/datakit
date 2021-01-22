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
type CancelAutoRenewalResourcesResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CancelAutoRenewalResourcesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CancelAutoRenewalResourcesResponse struct{}"
	}

	return strings.Join([]string{"CancelAutoRenewalResourcesResponse", string(data)}, " ")
}
