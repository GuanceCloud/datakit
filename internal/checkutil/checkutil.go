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
