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

// Request Object
type CreateBareMetalServersRequest struct {
	Body *CreateBaremetalServersBody `json:"body,omitempty"`
}

func (o CreateBareMetalServersRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateBareMetalServersRequest struct{}"
	}

	return strings.Join([]string{"CreateBareMetalServersRequest", string(data)}, " ")
}
