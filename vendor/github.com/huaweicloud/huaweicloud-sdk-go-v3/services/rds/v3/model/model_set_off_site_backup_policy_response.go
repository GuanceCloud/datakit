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
type SetOffSiteBackupPolicyResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o SetOffSiteBackupPolicyResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SetOffSiteBackupPolicyResponse struct{}"
	}

	return strings.Join([]string{"SetOffSiteBackupPolicyResponse", string(data)}, " ")
}
