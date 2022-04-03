// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package worker

import (
	"errors"
	"testing"
)

func TestErr(t *testing.T) {
	errr := func() error {
		return ErrTaskChNotReady
	}()
	if errors.Is(ErrTaskChNotReady, errr) != true {
		t.Error(errr)
	}

	closedErr := errors.New("pipeline task channal is closed")
	if errors.Is(ErrTaskChClosed, closedErr) == true {
		t.Error(closedErr)
	}
}
