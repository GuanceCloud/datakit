/*
 * SWR
 *
 * SWR API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type Certification struct {
	Cert *AuthInfo `json:"cert"`
}

func (o Certification) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Certification struct{}"
	}

	return strings.Join([]string{"Certification", string(data)}, " ")
}
