package inputs

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

type testinput struct {
	name string
	x    interface{}
}

func (ti *testinput) Catalog() string      { return "" }
func (ti *testinput) SampleConfig() string { return "" }

func (ti *testinput) another() {
	_ = ti.x.(map[string]string)["xy"] // panic here: type error
	panic(fmt.Errorf("panic error"))   // panic here: panic get a error type arg
}

func (ti *testinput) Run() {
	time.Sleep(time.Second)
	l.Debug("try panic here...")
	ti.another()
}

func TestProtectRunningInput(t *testing.T) {
	MaxCrash = 1
	i := testinput{name: `test-input`}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		protectRunningInput(i.name, &inputInfo{input: &i})
	}()

	wg.Wait()

	l.Debugf("panics: %+#v", GetPanicCnt(i.name))
}
