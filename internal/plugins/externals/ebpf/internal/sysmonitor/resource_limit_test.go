//go:build linux
// +build linux

package sysmonitor

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseBytes(t *testing.T) {
	cases := [][2]any{
		{"2MB", float64(2 * 1000 * 1000)},
		{"2MiB", float64(2 * 1024 * 1024)},
		{"2GB", float64(2 * 1000 * 1000 * 1000)},
		{"2GiB", float64(2 * 1024 * 1024 * 1024)},
	}

	for _, v := range cases {
		t.Run(v[0].(string), func(t *testing.T) {
			r := GetBytes(v[0].(string))
			assert.Equal(t, v[1], r)
		})
	}
}

func TestMonitor(t *testing.T) {
	proc, err := SelfProcess()
	if err != nil {
		t.Fatal(err)
	}

	// cpu
	{
		res := NewResLimiter(proc, 1, 0, 0)
		assert.Equal(t, false, res.overResLimit())

		ch := make(chan struct{})
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					_ = 1.0 + 12.0
					_ = 111.0 + 12.0
					_ = 1.0 + 111.0
					select {
					case <-ch:
						return
					default:
					}
				}
			}()
		}
		time.Sleep(time.Second)
		assert.Equal(t, true, res.overResLimit())
		close(ch)
		wg.Wait()
	}
	// mem
	{
		runtime.GC()

		m, err := proc.MemoryInfo()
		if err != nil {
			t.Fatal(err)
		}
		res := NewResLimiter(proc, 0, float64(m.RSS)+1e6, 0)
		assert.Equal(t, false, res.overResLimit())
		a := make([]byte, 1e8)
		assert.Equal(t, true, res.overResLimit())
		_ = a
	}
}
