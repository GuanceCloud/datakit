/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ChangeFailoverStrategyResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o ChangeFailoverStrategyResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ChangeFailoverStrategyResponse struct{}"
	}

	return strings.Join([]string{"ChangeFailoverStrategyResponse", string(data)}, " ")
}
