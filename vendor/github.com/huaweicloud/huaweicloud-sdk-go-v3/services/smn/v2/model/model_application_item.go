/*
 * SMN
 *
 * SMN Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ApplicationItem struct {
	// 创建application的名字。
	Name string `json:"name"`
	// 应用平台。
	Platform string `json:"platform"`
	// 创建application的时间时间格式为UTC时间，YYYY-MM-DDTHH:MM:SSZ。
	CreateTime string `json:"create_time"`
	// Application的唯一资源标识。
	ApplicationUrn string `json:"application_urn"`
	// Application的唯一标识ID。
	ApplicationId string `json:"application_id"`
	// 应用平台是否启用。
	Enabled string `json:"enabled"`
	// 苹果证书过期时间APNS、APNS_SANDBOX平台特有属性时间格式为UTC时间，YYYY-MM-DDTHH:MM:SSZ。
	AppleCertificateExpirationDate *string `json:"apple_certificate_expiration_date,omitempty"`
}

func (o ApplicationItem) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ApplicationItem struct{}"
	}

	return strings.Join([]string{"ApplicationItem", string(data)}, " ")
}
