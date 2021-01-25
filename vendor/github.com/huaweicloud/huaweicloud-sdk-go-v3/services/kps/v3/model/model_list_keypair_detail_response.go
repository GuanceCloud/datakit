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
type ListKeypairDetailResponse struct {
	Keypair        *KeypairDetail `json:"keypair,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o ListKeypairDetailResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListKeypairDetailResponse struct{}"
	}

	return strings.Join([]string{"ListKeypairDetailResponse", string(data)}, " ")
}
