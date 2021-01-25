/*
 * kms
 *
 * KMS v1.0 API, open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ShowUserQuotasResponse struct {
	Quotas         *Quotas `json:"quotas,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ShowUserQuotasResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowUserQuotasResponse struct{}"
	}

	return strings.Join([]string{"ShowUserQuotasResponse", string(data)}, " ")
}
