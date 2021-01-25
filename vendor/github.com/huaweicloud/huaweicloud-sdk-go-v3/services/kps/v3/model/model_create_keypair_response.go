/*
 * kps
 *
 * kps v3 版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type CreateKeypairResponse struct {
	Keypair        *CreateKeypairResp `json:"keypair,omitempty"`
	HttpStatusCode int                `json:"-"`
}

func (o CreateKeypairResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateKeypairResponse struct{}"
	}

	return strings.Join([]string{"CreateKeypairResponse", string(data)}, " ")
}
