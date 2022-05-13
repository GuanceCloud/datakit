// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package checkutil contains check utils
package checkutil

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func CheckConditionExit(f func() bool) {
	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for {
		if f() {
			return
		}

		select {
		case <-tick.C:

		case <-datakit.Exit.Wait():
			return
		}
	}
}
