/*
 * AS
 *
 * 弹性伸缩API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ShowResourceQuotaResponse struct {
	Quotas         *AllQuotas `json:"quotas,omitempty"`
	HttpStatusCode int        `json:"-"`
}

func (o ShowResourceQuotaResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowResourceQuotaResponse struct{}"
	}

	return strings.Join([]string{"ShowResourceQuotaResponse", string(data)}, " ")
}
