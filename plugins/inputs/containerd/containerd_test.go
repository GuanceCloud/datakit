// +build linux

package containerd

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestStart(t *testing.T) {

	var e = Containerd{
		Config: Config{
			Subscribes: []Subscribe{
				Subscribe{
					HostPath:    "/run/containerd/containerd.sock",
					Namespace:   "moby",
					IDList:      []string{"05da29442b461c6c97dca2838486a92a35ee6b1739240570b59c09504a300bb7"},
					Cycle:       3,
					Measurement: "measurement_111",
				},
			},
		},
	}

	e.wg = new(sync.WaitGroup)

	for _, sub := range e.Config.Subscribes {
		e.wg.Add(1)
		s := sub
		stream := newStream(&s, nil)
		fmt.Println(s)
		if err := stream.processMetrics(); err != nil {
			panic(err)
		}
		go stream.start(e.wg)
	}

	time.Sleep(10 * time.Second)
}
