package mongodb

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestStart(t *testing.T) {

	var m = Mongodb{
		Config: Config{
			Subscribes: []Subscribe{
				Subscribe{
					MongodbURL:  "mongodb://10.100.64.106:30001",
					Database:    "db123",
					Collection:  "tb123",
					Measurement: "mea123",
					Tags:        []string{"/name"},
					Fields:      map[string]string{"/age": "int", "/ob/tt": "int"},
				},
			},
		},
	}

	m.wg = new(sync.WaitGroup)

	for _, sub := range m.Config.Subscribes {
		m.wg.Add(1)
		fmt.Printf("%#v\n", sub)
		stream := newStream(&sub)
		panic(stream.start(m.wg))
	}
	time.Sleep(10 * time.Second)

}
