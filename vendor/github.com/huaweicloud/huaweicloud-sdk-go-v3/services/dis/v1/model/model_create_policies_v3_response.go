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

// Response Object
type CreatePoliciesV3Response struct {
	HttpStatusCode int `json:"-"`
}

func (o CreatePoliciesV3Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreatePoliciesV3Response struct{}"
	}

	return strings.Join([]string{"CreatePoliciesV3Response", string(data)}, " ")
}
