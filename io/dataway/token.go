// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"fmt"
	"strings"
)

var tokenFormatMap = map[string]int{
	"token": 32,
	"tkn":   32,
	"tokn":  24,
}

func CheckToken(token string) error {
	parts := strings.Split(token, "_")

	if len(parts) == 2 {
		prefix := parts[0]
		tokenVal := parts[1]

		if tokenLen, ok := tokenFormatMap[prefix]; ok {
			if len(tokenVal) != tokenLen {
				return fmt.Errorf("invalid token format, expect token length %d, but got %d",
					tokenLen+len(prefix)+1, len(tokenVal)+len(prefix)+1)
			} else {
				return nil
			}
		}
	}

	return fmt.Errorf("token invalid format")
}
