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

type ApiLink struct {
	// API的URL地址。
	Href *string `json:"href,omitempty"`
	// 默认值self。
	Rel *string `json:"rel,omitempty"`
}

func (o ApiLink) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ApiLink struct{}"
	}

	return strings.Join([]string{"ApiLink", string(data)}, " ")
}
