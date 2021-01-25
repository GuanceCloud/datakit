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
type PayOrdersResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o PayOrdersResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PayOrdersResponse struct{}"
	}

	return strings.Join([]string{"PayOrdersResponse", string(data)}, " ")
}
