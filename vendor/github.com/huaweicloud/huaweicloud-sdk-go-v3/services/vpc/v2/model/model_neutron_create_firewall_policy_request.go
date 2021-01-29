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
type NeutronCreateFirewallPolicyRequest struct {
	Body *NeutronCreateFirewallPolicyRequestBody `json:"body,omitempty"`
}

func (o NeutronCreateFirewallPolicyRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronCreateFirewallPolicyRequest struct{}"
	}

	return strings.Join([]string{"NeutronCreateFirewallPolicyRequest", string(data)}, " ")
}
