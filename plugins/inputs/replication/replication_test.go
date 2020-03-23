package replication

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var r = Replication{
		Config: Config{
			Subscribes: []Subscribe{
				Subscribe{
					Host:        "10.100.64.106",
					Port:        25432,
					Database:    "testdb",
					User:        "rep_name",
					Password:    "Sql123456",
					Table:       "tb",
					SlotName:    "slot_for_kafka",
					Measurement: "hello",
					Tags:        nil,
					Fields:      []string{"name"},
				},
			},
		},
		ctx:    ctx,
		cancel: cancel,
	}

	r.wg = new(sync.WaitGroup)

	for _, sub := range r.Config.Subscribes {
		r.wg.Add(1)
		stream := newStream(&sub)
		stream.start(&r, r.ctx, r.wg)
	}

	time.Sleep(5 * time.Second)

}
