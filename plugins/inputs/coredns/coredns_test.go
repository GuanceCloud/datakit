package coredns

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestStart(t *testing.T) {

	var e = Coredns{
		Config: Config{
			Subscribes: []Subscribe{
				Subscribe{
					CorednsHost: "10.100.64.106",
					CorednsPort: 29153,
					Cycle:       3,
					Measurement: "coredns_measurement",
				},
			},
		},
	}

	e.wg = new(sync.WaitGroup)

	for _, sub := range e.Config.Subscribes {
		e.wg.Add(1)
		fmt.Printf("%#v\n", sub)
		stream := newStream(&sub, nil)
		panic(stream.start(e.wg))
	}

	time.Sleep(10 * time.Second)

}
