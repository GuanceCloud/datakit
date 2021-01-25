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

// Request Object
type ListKeypairDetailRequest struct {
	KeypairName string `json:"keypair_name"`
}

func (o ListKeypairDetailRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListKeypairDetailRequest struct{}"
	}

	return strings.Join([]string{"ListKeypairDetailRequest", string(data)}, " ")
}
