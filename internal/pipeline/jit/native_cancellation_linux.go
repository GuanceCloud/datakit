// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// Copyright 2026-present Guance, Inc.

package jit

import (
	"context"
	"errors"
)

const maxIdleCancellations = 64

type nativeCancellation struct {
	token uintptr
	done  chan struct{}
}

func (n *nativeRuntime) watchCancellation(ctx context.Context) (uintptr, func(), error) {
	n.cancelPoolMu.Lock()
	var lease *nativeCancellation
	if size := len(n.cancelPool); size > 0 {
		lease = n.cancelPool[size-1]
		n.cancelPool[size-1] = nil
		n.cancelPool = n.cancelPool[:size-1]
	}
	n.cancelPoolMu.Unlock()
	if lease == nil {
		lease = &nativeCancellation{done: make(chan struct{})}
		if status := n.cancelCreate(&lease.token); status != 0 || lease.token == 0 {
			return 0, nil, nativeError("create cancellation", status, nil)
		}
	}
	stop := context.AfterFunc(ctx, func() {
		defer close(lease.done)
		n.cancelRequest(lease.token)
	})
	return lease.token, func() {
		if !stop() {
			// A started callback must finish before token destruction and dlclose.
			// Its token is permanently cancelled and must never be recycled.
			<-lease.done
			n.cancelDestroy(lease.token)
			return
		}
		// stop=true guarantees the old callback cannot touch this token again.
		n.cancelPoolMu.Lock()
		if len(n.cancelPool) < maxIdleCancellations {
			n.cancelPool = append(n.cancelPool, lease)
			lease = nil
		}
		n.cancelPoolMu.Unlock()
		if lease != nil {
			n.cancelDestroy(lease.token)
		}
	}, nil
}

func (n *nativeRuntime) closeCancellationPool() error {
	n.cancelPoolMu.Lock()
	defer n.cancelPoolMu.Unlock()
	var err error
	for _, lease := range n.cancelPool {
		if status := n.cancelDestroy(lease.token); status != 0 {
			err = errors.Join(err, nativeError("destroy idle cancellation", status, nil))
		}
	}
	n.cancelPool = nil
	return err
}
