// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package endpoint

import (
	"errors"
	"strings"
)

var (
	ErrWritePoints4XX    = errors.New("write point 4xx")
	ErrRequestTerminated = errors.New("no response and request maybe terminated")
	ErrDirtyUpload       = errors.New("dirty upload data")
)

func isDirtyUploadError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "invalid header field") ||
		strings.Contains(msg, "invalid header value") ||
		strings.Contains(msg, "invalid header name") ||
		strings.Contains(msg, "unsupported protocol")
}
