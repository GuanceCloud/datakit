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

//
type NovaListKeypairsResult struct {
	Keypair *NovaSimpleKeypair `json:"keypair"`
}

func (o NovaListKeypairsResult) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NovaListKeypairsResult struct{}"
	}

	return strings.Join([]string{"NovaListKeypairsResult", string(data)}, " ")
}
