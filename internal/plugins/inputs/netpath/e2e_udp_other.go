// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !linux
// +build !linux

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netpath

import (
	"context"
	"errors"
)

func runUDPE2E(t task) e2eResult {
	return runUDPE2EContext(context.Background(), t)
}

func runUDPE2EContext(_ context.Context, _ task) e2eResult {
	return e2eResult{
		protocol: protocolUDP,
		err:      errors.New("udp e2e probe is only supported on linux"),
	}
}
