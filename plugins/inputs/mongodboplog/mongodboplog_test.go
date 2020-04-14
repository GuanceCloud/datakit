package mongodboplog

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestStart(t *testing.T) {

	var m = Mongodboplog{
		Config: Config{
			Subscribes: []Subscribe{
				Subscribe{
					MongodbURL:  "mongodb://10.100.64.106:30001",
					Database:    "db123",
					Collection:  "tb123",
					Measurement: "test_mea",
					Tags:        []string{"/name", "/address/home"},
					Fields: map[string]string{
						"/age":            "int",
						"/address/school": "string",
						"/ob2\\[0\\]":     "int",
						"/ob2[1]/name":    "string",
					},
				},
			},
		},
	}

	m.wg = new(sync.WaitGroup)

	for _, sub := range m.Config.Subscribes {
		m.wg.Add(1)
		fmt.Printf("%#v\n", sub)
		stream := newStream(&sub, nil)
		panic(stream.start(m.wg))
	}
	time.Sleep(10 * time.Second)

}
