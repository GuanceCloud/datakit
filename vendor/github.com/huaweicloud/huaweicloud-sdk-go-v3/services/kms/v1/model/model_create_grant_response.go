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
type CreateGrantResponse struct {
	// 授权ID，64字节。
	GrantId        *string `json:"grant_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CreateGrantResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateGrantResponse struct{}"
	}

	return strings.Join([]string{"CreateGrantResponse", string(data)}, " ")
}
