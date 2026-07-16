// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !linux && !windows

package netpath

import "fmt"

func validateProtocolForPlatform(protocol string) error {
	if protocol == protocolTCP || protocol == protocolUDP {
		return fmt.Errorf("unsupported netpath protocol %q on this platform", protocol)
	}
	return nil
}
