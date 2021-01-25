/*
 * VPC
 *
 * VPC Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type CreateSecurityGroupRuleRequest struct {
	Body *CreateSecurityGroupRuleRequestBody `json:"body,omitempty"`
}

func (o CreateSecurityGroupRuleRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateSecurityGroupRuleRequest struct{}"
	}

	return strings.Join([]string{"CreateSecurityGroupRuleRequest", string(data)}, " ")
}
