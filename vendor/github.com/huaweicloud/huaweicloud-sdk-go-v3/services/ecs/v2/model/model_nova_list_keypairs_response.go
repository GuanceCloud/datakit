/*
 * ECS
 *
 * ECS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type NovaListKeypairsResponse struct {
	// 密钥信息列表。
	Keypairs       *[]NovaListKeypairsResult `json:"keypairs,omitempty"`
	HttpStatusCode int                       `json:"-"`
}

func (o NovaListKeypairsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NovaListKeypairsResponse struct{}"
	}

	return strings.Join([]string{"NovaListKeypairsResponse", string(data)}, " ")
}
