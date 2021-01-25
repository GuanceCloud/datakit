/*
 * CCE
 *
 * CCE开放API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ListAddonTemplatesRequest struct {
	ContentType       string  `json:"Content-Type"`
	AddonTemplateName *string `json:"addon_template_name,omitempty"`
}

func (o ListAddonTemplatesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListAddonTemplatesRequest struct{}"
	}

	return strings.Join([]string{"ListAddonTemplatesRequest", string(data)}, " ")
}
