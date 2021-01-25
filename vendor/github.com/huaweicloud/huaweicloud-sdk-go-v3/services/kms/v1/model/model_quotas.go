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

type Quotas struct {
	// 资源配额列表，详情请参见Resources
	Resources *[]Resources `json:"resources,omitempty"`
}

func (o Quotas) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Quotas struct{}"
	}

	return strings.Join([]string{"Quotas", string(data)}, " ")
}
