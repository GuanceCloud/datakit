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

// Request Object
type ShowResourceQuotaRequest struct {
}

func (o ShowResourceQuotaRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowResourceQuotaRequest struct{}"
	}

	return strings.Join([]string{"ShowResourceQuotaRequest", string(data)}, " ")
}
