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
type ShowPrivateipRequest struct {
	PrivateipId string `json:"privateip_id"`
}

func (o ShowPrivateipRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowPrivateipRequest struct{}"
	}

	return strings.Join([]string{"ShowPrivateipRequest", string(data)}, " ")
}
