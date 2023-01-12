// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"errors"
	"io"
	"strings"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestLimitReader_Close(t *testing.T) {
	r := io.NopCloser(strings.NewReader("hello world!!!!!"))

	lr := newLimitReader(r, 10)

	c, err := io.ReadAll(lr)

	t.Log(err, len(c), string(c), errors.Is(err, errLimitReader))

	testutil.NotOk(t, err, "")
	testutil.Equals(t, 10, len(c))
	testutil.Equals(t, true, errors.Is(err, errLimitReader))
}
