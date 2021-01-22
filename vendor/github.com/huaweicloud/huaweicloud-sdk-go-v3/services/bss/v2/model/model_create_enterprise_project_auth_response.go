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
type CreateEnterpriseProjectAuthResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CreateEnterpriseProjectAuthResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateEnterpriseProjectAuthResponse struct{}"
	}

	return strings.Join([]string{"CreateEnterpriseProjectAuthResponse", string(data)}, " ")
}
