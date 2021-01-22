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
type DeletePostalResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeletePostalResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeletePostalResponse struct{}"
	}

	return strings.Join([]string{"DeletePostalResponse", string(data)}, " ")
}
