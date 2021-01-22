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
type CreateSubEnterpriseAccountResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CreateSubEnterpriseAccountResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateSubEnterpriseAccountResponse struct{}"
	}

	return strings.Join([]string{"CreateSubEnterpriseAccountResponse", string(data)}, " ")
}
