/*
 * FunctionGraph
 *
 * API v2
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListQuotasResponse struct {
	Quotas         *ListQuotasResult `json:"quotas,omitempty"`
	HttpStatusCode int               `json:"-"`
}

func (o ListQuotasResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListQuotasResponse struct{}"
	}

	return strings.Join([]string{"ListQuotasResponse", string(data)}, " ")
}
