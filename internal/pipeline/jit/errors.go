// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package jit

import (
	"context"
	"errors"
	"fmt"
)

type NativeCallError struct {
	Operation string
	Status    int32
	Detail    string
}

func (e *NativeCallError) Error() string {
	if e.Detail == "" {
		return fmt.Sprintf("JIT %s returned status %d", e.Operation, e.Status)
	}
	return fmt.Sprintf("JIT %s returned status %d: %s", e.Operation, e.Status, e.Detail)
}

type ProtocolError struct{ Cause error }

func (e *ProtocolError) Error() string { return e.Cause.Error() }
func (e *ProtocolError) Unwrap() error { return e.Cause }

// EngineFault never classifies errors by their text. Status 3 is the ABI's
// process failure (including resource rejection), not proof of engine damage.
func EngineFault(err error) (fault, runtimeWide bool) {
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false, false
	}
	var protocol *ProtocolError
	if errors.As(err, &protocol) {
		return true, false
	}
	var native *NativeCallError
	if errors.As(err, &native) {
		switch native.Status {
		case 2, 4:
			return true, true // ABI mismatch or caught native panic.
		case 0, 1, 3:
			return false, false
		default:
			return true, true // Unknown status is a broken ABI contract.
		}
	}
	return false, false
}

func classifyProcessError(err error) error { //nolint:unused // Used by the pipeline_jit build on Linux.
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	var native *NativeCallError
	var protocol *ProtocolError
	if errors.As(err, &native) || errors.As(err, &protocol) {
		return err
	}
	return &ProtocolError{Cause: err}
}
