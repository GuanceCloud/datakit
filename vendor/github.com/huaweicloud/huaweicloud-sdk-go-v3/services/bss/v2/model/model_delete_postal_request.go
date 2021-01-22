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

// Request Object
type DeletePostalRequest struct {
	AddressId string `json:"address_id"`
}

func (o DeletePostalRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeletePostalRequest struct{}"
	}

	return strings.Join([]string{"DeletePostalRequest", string(data)}, " ")
}
