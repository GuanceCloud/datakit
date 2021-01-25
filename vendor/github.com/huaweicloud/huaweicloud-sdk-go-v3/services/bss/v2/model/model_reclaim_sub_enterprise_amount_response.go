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
type ReclaimSubEnterpriseAmountResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o ReclaimSubEnterpriseAmountResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ReclaimSubEnterpriseAmountResponse struct{}"
	}

	return strings.Join([]string{"ReclaimSubEnterpriseAmountResponse", string(data)}, " ")
}
