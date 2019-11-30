package utils

import (
	"fmt"
	"sync"
	"time"
)

func SizeToName(size int64) string {
	units := []string{"B", "KB", "MB", "GB"}
	i := 0
	for size >= 1024 {
		size /= 1024
		i++
	}

	if i > len(units)-1 {
		i = len(units) - 1
	}

	return fmt.Sprintf("%d%s", size, units[i])
}

type rateLimiter struct {
	C    chan bool
	rate time.Duration
	n    int

	shutdown chan bool
	wg       sync.WaitGroup
}

func NewRateLimiter(n int, rate time.Duration) *rateLimiter {
	r := &rateLimiter{
		C:        make(chan bool),
		rate:     rate,
		n:        n,
		shutdown: make(chan bool),
	}
	r.wg.Add(1)
	go r.limiter()
	return r
}

func (r *rateLimiter) Stop() {
	close(r.shutdown)
	r.wg.Wait()
	close(r.C)
}

func (r *rateLimiter) limiter() {
	defer r.wg.Done()
	ticker := time.NewTicker(r.rate)
	defer ticker.Stop()
	counter := 0
	for {
		select {
		case <-r.shutdown:
			return
		case <-ticker.C:
			counter = 0
		default:
			if counter < r.n {
				select {
				case r.C <- true:
					counter++
				case <-r.shutdown:
					return
				}
			} else {
				time.Sleep(time.Millisecond * 5)
			}
		}
	}
}
