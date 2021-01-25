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
type ShowTenantQuotaRequest struct {
}

func (o ShowTenantQuotaRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowTenantQuotaRequest struct{}"
	}

	return strings.Join([]string{"ShowTenantQuotaRequest", string(data)}, " ")
}
