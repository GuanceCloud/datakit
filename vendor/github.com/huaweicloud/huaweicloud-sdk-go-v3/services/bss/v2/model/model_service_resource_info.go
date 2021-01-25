/*
 * BSS
 *
 * Business Support System API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ServiceResourceInfo struct {
	BasicInfo *ResourceBasicInfo `json:"basic_info,omitempty"`
}

func (o ServiceResourceInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ServiceResourceInfo struct{}"
	}

	return strings.Join([]string{"ServiceResourceInfo", string(data)}, " ")
}
