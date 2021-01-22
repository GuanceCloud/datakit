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
type UpdateSubEnterpriseAmountResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o UpdateSubEnterpriseAmountResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateSubEnterpriseAmountResponse struct{}"
	}

	return strings.Join([]string{"UpdateSubEnterpriseAmountResponse", string(data)}, " ")
}
