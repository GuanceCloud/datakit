// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kingbase

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mitchellh/mapstructure"
)

func mapToStruct(m map[string]any, dest any) error {
	// map key 统一小写
	lowerMap := make(map[string]any, len(m))
	for k, v := range m {
		// lowerMap[strings.ToLower(k)] = v
		lk := strings.ToLower(k)
		switch vv := v.(type) {
		case nil:
			// leave as nil (mapstructure + WeaklyTypedInput 会把 nil -> zero value)
			lowerMap[lk] = nil
		case []byte:
			// 在有些 driver/场景会以 []byte 返回，先转成 string
			lowerMap[lk] = string(vv)
		default:
			lowerMap[lk] = vv
		}
	}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          "db",
		Result:           dest,
		WeaklyTypedInput: true,
		MatchName:        strings.EqualFold,
	})
	if err != nil {
		return err
	}

	return decoder.Decode(lowerMap)
}

func extractVersion(vs string) (string, error) {
	re := regexp.MustCompile(`V(\d+)R(\d+)`)
	matches := re.FindStringSubmatch(vs)
	if len(matches) < 3 {
		return "", fmt.Errorf("invalid version format: %s", vs)
	}

	major := strings.TrimLeft(matches[1], "0")
	if major == "" {
		major = "0"
	}
	minor := strings.TrimLeft(matches[2], "0")
	if minor == "" {
		minor = "0"
	}

	return fmt.Sprintf("V%sR%s", major, minor), nil
}
