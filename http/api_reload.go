// +build !windows

package http

import (
	"time"
)

func RestartHttpServer() {
	reload = time.Now()
	reloadCnt++

	HttpStart()
}
