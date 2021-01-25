/*
 * RMS
 *
 * Resource Manager Api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ListBuiltInPolicyDefinitionsRequest struct {
	XLanguage *string `json:"X-Language,omitempty"`
}

func (o ListBuiltInPolicyDefinitionsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListBuiltInPolicyDefinitionsRequest struct{}"
	}

	return strings.Join([]string{"ListBuiltInPolicyDefinitionsRequest", string(data)}, " ")
}
