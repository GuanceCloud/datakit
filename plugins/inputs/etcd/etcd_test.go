package etcd

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestStart(t *testing.T) {

	var e = Etcd{
		Config: Config{
			Subscribes: []Subscribe{
				Subscribe{
					EtcdHost:    "10.100.64.106",
					EtcdPort:    32379,
					Cycle:       3,
					Measurement: "etcd_measurement",
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
