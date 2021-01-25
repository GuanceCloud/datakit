/*
 * DevStar
 *
 * DevStar API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type TemplateViewHistory struct {
	// 模板的id
	TemplateId string `json:"template_id"`
	// 模板的名称
	TemplateTitle string `json:"template_title"`
}

func (o TemplateViewHistory) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TemplateViewHistory struct{}"
	}

	return strings.Join([]string{"TemplateViewHistory", string(data)}, " ")
}
