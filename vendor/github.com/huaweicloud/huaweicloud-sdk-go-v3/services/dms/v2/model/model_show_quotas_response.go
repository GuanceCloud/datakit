/*
 * DMS
 *
 * DMS Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ShowQuotasResponse struct {
	Quotas         *ShowQuotasRespQuotas `json:"quotas,omitempty"`
	HttpStatusCode int                   `json:"-"`
}

func (o ShowQuotasResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowQuotasResponse struct{}"
	}

	return strings.Join([]string{"ShowQuotasResponse", string(data)}, " ")
}
