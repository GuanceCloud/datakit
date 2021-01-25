/*
 * CCE
 *
 * CCE开放API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type DeleteAddonInstanceResponse struct {
	Body           *string `json:"body,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o DeleteAddonInstanceResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteAddonInstanceResponse struct{}"
	}

	return strings.Join([]string{"DeleteAddonInstanceResponse", string(data)}, " ")
}
