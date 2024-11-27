// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package lsblk

import (
	"testing"
)

func TestCollectLsblkInfo(t *testing.T) {
	ipt := defaultInput()

	if _, err := ipt.collectLsblkInfo(); err != nil {
		t.Errorf("Failed to collect lsblk info: %v", err)
	}
	// fmt.Fprintln(os.Stdout, devices)
}
