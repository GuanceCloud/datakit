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

type ListApplicationAttributesResponseBodyAttributes struct {
	// 应用平台是否启用。
	Enabled string `json:"enabled"`
	// 苹果证书过期时间，APNS、APNS_SANDBOX平台特有属性 时间格式为UTC时间，YYYY-MM-DDTHH:MM:SSZ。
	AppleCertificateExpirationDate *string `json:"apple_certificate_expiration_date,omitempty"`
}

func (o ListApplicationAttributesResponseBodyAttributes) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListApplicationAttributesResponseBodyAttributes struct{}"
	}

	return strings.Join([]string{"ListApplicationAttributesResponseBodyAttributes", string(data)}, " ")
}
