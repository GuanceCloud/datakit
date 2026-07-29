// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netpath

import "sync"

var registerOnce sync.Once //nolint:gochecknoglobals

// Register enables the standalone netpath input and its process-wide
// integrations. Probe-only consumers should import this package without
// calling Register so they do not alter the input variant matrix.
func Register() {
	registerOnce.Do(func() {
		registerInput()
		registerHTTPRouteMatcher()
		registerMetrics()
	})
}
