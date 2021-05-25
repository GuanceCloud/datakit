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
type ShowBuiltInPolicyDefinitionRequest struct {
	PolicyDefinitionId string  `json:"policy_definition_id"`
	XLanguage          *string `json:"X-Language,omitempty"`
}

func (o ShowBuiltInPolicyDefinitionRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowBuiltInPolicyDefinitionRequest struct{}"
	}

	return strings.Join([]string{"ShowBuiltInPolicyDefinitionRequest", string(data)}, " ")
}
