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

// Request Object
type DeletePrivateipRequest struct {
	PrivateipId string `json:"privateip_id"`
}

func (o DeletePrivateipRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeletePrivateipRequest struct{}"
	}

	return strings.Join([]string{"DeletePrivateipRequest", string(data)}, " ")
}
