package flink

import (
	"strconv"
	"sync"
	"testing"
)

func TestHash(t *testing.T) {

	var dblist = Flinks{m: make(map[string]map[string]string), mut: &sync.RWMutex{}}

	wg := sync.WaitGroup{}
	for idx := 1; idx < 10; idx++ {
		wg.Add(1)
		go func(i int) {
			k := strconv.Itoa(i*100 + i*10 + i)
			dblist.Store(k, map[string]string{k: k})
			wg.Done()
		}(idx)
	}

	for idx := 0; idx < 5; idx++ {
		wg.Add(1)
		go func(i int) {
			k := strconv.Itoa(i*100 + i*10 + i)
			for cnt := 0; cnt < 1000; cnt++ {
				v, ok := dblist.Load(k)
				t.Logf("%s: %v:%v\n", k, v, ok)
			}
			wg.Done()
		}(idx)
	}

	wg.Wait()
}
