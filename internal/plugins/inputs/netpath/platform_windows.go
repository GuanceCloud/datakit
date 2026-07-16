// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows

package netpath

import "fmt"

func validateProtocolForPlatform(protocol string) error {
	return fmt.Errorf("unsupported netpath protocol %q on windows", protocol)
}
