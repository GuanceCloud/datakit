/*
 * BSS
 *
 * Business Support System API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type CreateEnterpriseProjectAuthRequest struct {
}

func (o CreateEnterpriseProjectAuthRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateEnterpriseProjectAuthRequest struct{}"
	}

	return strings.Join([]string{"CreateEnterpriseProjectAuthRequest", string(data)}, " ")
}
