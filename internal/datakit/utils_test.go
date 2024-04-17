// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package datakit

import (
	T "testing"
	"time"

	tu "github.com/GuanceCloud/cliutils/testutil"
)

func TestDuration(t *T.T) {
	d := Duration{Duration: time.Second}
	tu.Equals(t, "1s", d.UnitString(time.Second))
	tu.Equals(t, "1000000000ns", d.UnitString(time.Nanosecond))
	tu.Equals(t, "1000000mics", d.UnitString(time.Microsecond))
	tu.Equals(t, "1000ms", d.UnitString(time.Millisecond))
	tu.Equals(t, "0m", d.UnitString(time.Minute))
	tu.Equals(t, "0h", d.UnitString(time.Hour))
}
