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
type DeleteHotkeyScanTaskResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteHotkeyScanTaskResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteHotkeyScanTaskResponse struct{}"
	}

	return strings.Join([]string{"DeleteHotkeyScanTaskResponse", string(data)}, " ")
}
