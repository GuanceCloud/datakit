/*
 * DIS
 *
 * DIS v1 API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ListPoliciesV3Request struct {
	StreamName string `json:"stream_name"`
}

func (o ListPoliciesV3Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPoliciesV3Request struct{}"
	}

	return strings.Join([]string{"ListPoliciesV3Request", string(data)}, " ")
}
