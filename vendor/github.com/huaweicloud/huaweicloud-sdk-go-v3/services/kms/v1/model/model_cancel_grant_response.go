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
type CancelGrantResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CancelGrantResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CancelGrantResponse struct{}"
	}

	return strings.Join([]string{"CancelGrantResponse", string(data)}, " ")
}
