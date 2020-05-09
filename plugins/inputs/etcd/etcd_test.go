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
					CacertFile:  "",
					CertFile:    "",
					KeyFile:     "",
					Cycle:       3,
					Measurement: "etcd_measurement",
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
		go stream.start(e.wg)
	}

	time.Sleep(10 * time.Second)
}
