/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListBareMetalServerDetailsResponse struct {
	Server         *ServerDetails `json:"server,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o ListBareMetalServerDetailsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListBareMetalServerDetailsResponse struct{}"
	}

	return strings.Join([]string{"ListBareMetalServerDetailsResponse", string(data)}, " ")
}
