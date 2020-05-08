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
					EtcdHost:    "172.16.0.43",
					EtcdPort:    2379,
					CacertFile:  "/Users/liguozhuang/etcdTLS/ca.crt",
					CertFile:    "/Users/liguozhuang/etcdTLS/peer.crt",
					KeyFile:     "/Users/liguozhuang/etcdTLS/peer.key",
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
