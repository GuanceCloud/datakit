// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows

package traceroute

import (
	"context"
	"net"
)

func trace(ctx context.Context, target net.IP, cfg options) (Result, error) {
	switch cfg.protocol {
	case ProtocolICMP:
		return traceICMP(ctx, target, cfg)
	case ProtocolUDP:
		return traceUDP(ctx, target, cfg)
	case ProtocolTCP:
		return traceTCP(ctx, target, cfg)
	default:
		return Result{}, nil
	}
}
