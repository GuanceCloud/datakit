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
type CancelSelfGrantResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CancelSelfGrantResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CancelSelfGrantResponse struct{}"
	}

	return strings.Join([]string{"CancelSelfGrantResponse", string(data)}, " ")
}
