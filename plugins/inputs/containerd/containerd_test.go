// +build linux

package containerd

import (
	"fmt"
	"sync"
	"testing"
	// "time"
)

func TestStart(t *testing.T) {

	var e = Containerd{
		Config: Config{
			Subscribes: []Subscribe{
				Subscribe{
					HostPath:    "/run/containerd/containerd.sock",
					Namespace:   "moby",
					IDList:      []string{"*"},
					Cycle:       5,
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

		fmt.Println(stream.points)
		stream.points = nil
	}

	// time.Sleep(10 * time.Second)
}
