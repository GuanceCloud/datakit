package socket

import (
	"fmt"
	"sync"
	"testing"
)

func TestInput_Run(t *testing.T) {
	i := &Input{
		DestURL:  []string{"tcp:47.110.144.10:443", "udp:1.1.1.1:5555", "udp:1.1.1.1546786:5555"},
		curTasks: map[string]*dialer{},
		wg:       sync.WaitGroup{},
	}

	if err := i.Collect(); err != nil {
		l.Warnf(err.Error())
	}

	for _, c := range i.collectCache {
		fmt.Println(c.LineProto())
	}
}
