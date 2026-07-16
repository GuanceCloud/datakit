// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !linux && !windows

package traceroute

import (
	"context"
	"errors"
	"net"
)

func traceTCP(context.Context, net.IP, options) (Result, error) {
	return Result{}, errors.New("TCP traceroute is only supported on linux")
}
