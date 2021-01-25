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

type CreateNamespaceRequestBody struct {
	// 组织名称
	Namespace string `json:"namespace"`
}

func (o CreateNamespaceRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateNamespaceRequestBody struct{}"
	}

	return strings.Join([]string{"CreateNamespaceRequestBody", string(data)}, " ")
}
