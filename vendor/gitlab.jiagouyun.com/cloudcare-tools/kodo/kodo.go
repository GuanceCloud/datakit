package kodo

import (
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
)

var (
	Exit = cliutils.NewSem()
	WG   = sync.WaitGroup{}
)
