package socket

import (
	"fmt"
	"sync"
	"testing"
)

func TestInput_Run(t *testing.T) {
	i := &Input{
		DestUrl:  []string{"tcp:47.110.144.10:443", "udp:1.1.1.1:5555", "udp:1.1.1.1546786:5555", "udp:127.0.0.1:5555"},
		curTasks: map[string]*dialer{},
		wg:       sync.WaitGroup{},
	}
	err := i.Collect()
	if err != nil {
		fmt.Println(err.Error())
	}

	for _, c := range i.collectCache {
		fmt.Println(c.LineProto())
	}
	//time.Sleep(20 * time.Second)
}
