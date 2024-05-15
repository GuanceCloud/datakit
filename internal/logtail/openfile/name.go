// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package openfile

import "strings"

func SplitFilenameFromKey(key string) string {
	arr := strings.Split(key, "::")
	if len(arr) > 0 {
		return arr[0]
	}
	return ""
}
