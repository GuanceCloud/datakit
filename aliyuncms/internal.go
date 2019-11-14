package aliyuncms

import (
	"bytes"
	"strconv"
	"sync"
	"time"
	"unicode"
)

type (
	Duration struct {
		time.Duration
	}
)

func (d *Duration) UnmarshalTOML(b []byte) error {
	var err error
	b = bytes.Trim(b, `'`)

	d.Duration, err = time.ParseDuration(string(b))
	if err == nil {
		return nil
	}

	if uq, err := strconv.Unquote(string(b)); err == nil && len(uq) > 0 {
		d.Duration, err = time.ParseDuration(uq)
		if err == nil {
			return nil
		}
	}

	sI, err := strconv.ParseInt(string(b), 10, 64)
	if err == nil {
		d.Duration = time.Second * time.Duration(sI)
		return nil
	}

	sF, err := strconv.ParseFloat(string(b), 64)
	if err == nil {
		d.Duration = time.Second * time.Duration(sF)
		return nil
	}

	return nil
}

func SnakeCase(in string) string {
	runes := []rune(in)
	length := len(runes)

	var out []rune
	for i := 0; i < length; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
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
