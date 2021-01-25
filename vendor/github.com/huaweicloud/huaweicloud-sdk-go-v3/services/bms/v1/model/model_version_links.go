/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// API的url地址
type VersionLinks struct {
	// API的url地址
	Href *string `json:"href,omitempty"`
	// API的url地址依赖
	Rel *string `json:"rel,omitempty"`
}

func (o VersionLinks) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "VersionLinks struct{}"
	}

	return strings.Join([]string{"VersionLinks", string(data)}, " ")
}
