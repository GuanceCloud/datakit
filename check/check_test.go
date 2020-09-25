package check

import (
	"sync"
	"testing"

	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/all"
)

func TestCheckDatakitToml(t *testing.T) {
	s := `
access_key_id = 'ak_xxxxxxxxxxxxx'
access_key_secret = 'sk____________'
region_id = 'rg_yyyyyyyyyyyyyy'
	`

	if err := CheckInputToml(`aliyunprice`, []byte(s)); err != nil {
		t.Fatal(err)
	}

	if err := CheckInputToml(`aws_billing`, []byte(s)); err == nil {
		t.Fatal("should error")
	} else {
		t.Log(err.Error())
	}
}

func TestInputToml(t *testing.T) {
	s := `
  ## Whether to report per-cpu stats or not
  percpu = true
  ## Whether to report total system cpu stats or not
  totalcpu = true
  ## If true, collect raw CPU time metrics.
  collect_cpu_time = false
  ## If true, compute and report the sum of all non-idle CPU states.
  report_active = false
	`

	if err := CheckInputToml("cpu", []byte(s)); err != nil {
		t.Fatal(err)
	}

	if err := CheckInputToml("mem", []byte(s)); err == nil { // bad
		t.Fatal(err)
	} else {
		t.Logf("check error: %s", err.Error())
	}

	// concurrently checking
	wg := sync.WaitGroup{}

	f := func() {
		if err := CheckInputToml("cpu", []byte(s)); err != nil {
			t.Fatal(err)
		}
	}

	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			f()
		}()
	}

	wg.Wait()
}
