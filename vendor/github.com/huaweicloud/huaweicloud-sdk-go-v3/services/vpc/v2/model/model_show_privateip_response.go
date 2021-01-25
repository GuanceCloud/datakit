/*
 * VPC
 *
 * VPC Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ShowPrivateipResponse struct {
	Privateip      *Privateip `json:"privateip,omitempty"`
	HttpStatusCode int        `json:"-"`
}

func (o ShowPrivateipResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowPrivateipResponse struct{}"
	}

	return strings.Join([]string{"ShowPrivateipResponse", string(data)}, " ")
}
