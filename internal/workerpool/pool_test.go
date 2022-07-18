package workerpool

import (
	"fmt"
	"log"
	"sync"
	"testing"
	"time"
)

func TestWorkerPool(t *testing.T) {
	wpool := NewWorkerPool(9000)
	if err := wpool.Start(16); err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	n, m := 100, 100
	wg := sync.WaitGroup{}
	wg.Add(n)
	start := time.Now()
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()

			for j := 0; j < m; j++ {
				job, err := NewJob(
					WithInput(fmt.Sprintf("goroutine %d, job %d\n", i, j)),
					WithTimeout(time.Second),
					WithProcess(func(input interface{}) (output interface{}) {
						log.Printf("start process input %d:%d\n", i, j)

						return fmt.Sprintf("finish process %d:%d\n", i, j)
					}),
					WithProcessCallback(func(input, output interface{}, cost time.Duration, isTimeout bool) {
						log.Printf("finish process and callback, input: %v output: %v cost: %dms isTimeout: %v\n", input, output, cost/time.Millisecond, isTimeout)
					}),
				)
				if err != nil {
					t.Errorf(err.Error())
					continue
				}

				if err = wpool.MoreJob(job); err != nil {
					log.Println(err.Error())
				}
			}
		}(i)
	}
	wg.Wait()

	log.Printf("send jobs finished, cost %dms\n", time.Since(start)/time.Millisecond)

	time.Sleep(1 * time.Second)
	wpool.Shutdown()
}
