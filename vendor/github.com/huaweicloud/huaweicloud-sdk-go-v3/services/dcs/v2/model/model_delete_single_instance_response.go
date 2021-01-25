/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type DeleteSingleInstanceResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteSingleInstanceResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteSingleInstanceResponse struct{}"
	}

	return strings.Join([]string{"DeleteSingleInstanceResponse", string(data)}, " ")
}
