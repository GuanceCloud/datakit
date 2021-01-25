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
type ShowTenantQuotaResponse struct {
	Absolute       *Absolute `json:"absolute,omitempty"`
	HttpStatusCode int       `json:"-"`
}

func (o ShowTenantQuotaResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowTenantQuotaResponse struct{}"
	}

	return strings.Join([]string{"ShowTenantQuotaResponse", string(data)}, " ")
}
