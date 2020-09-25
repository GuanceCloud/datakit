package telegraf_inputs

import (
	"runtime"
	"sync"
	"testing"
)

func TestCheckWinInputUnderWindows(t *testing.T) {
	s := `#[[inputs.win_services]]
# Reports information about Windows service status.
# Monitoring some services may require running Telegraf with administrator privileges.
# Names of the services to monitor. Leave empty to monitor all the available services on the host
service_names = [
	"LanmanServer",
	"TermService",
]`

	if runtime.GOOS == "windows" {
		if err := CheckTelegrafToml("win_services", []byte(s)); err != nil {
			t.Fatal("should be `not-found-input`")
		}
	}
}

func TestCheckWinInputUnderLinux(t *testing.T) {
	s := `#[[inputs.win_services]]
# Reports information about Windows service status.
# Monitoring some services may require running Telegraf with administrator privileges.
# Names of the services to monitor. Leave empty to monitor all the available services on the host
service_names = [
	"LanmanServer",
	"TermService",
]`

	if runtime.GOOS != "windows" {
		if err := CheckTelegrafToml("win_services", []byte(s)); err == nil {
			t.Fatal("should be `not-found-input`")
		} else {
			t.Log(err)
		}
	}
}

func TestCheckTelegrafToml(t *testing.T) {
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
	if err := CheckTelegrafToml("cpu", []byte(s)); err != nil {
		t.Fatal(err)
	}

	if err := CheckTelegrafToml("mem", []byte(s)); err == nil { // bad
		t.Fatal(err)
	} else {
		t.Logf("check error: %s", err.Error())
	}

	// concurrently checking
	wg := sync.WaitGroup{}

	f := func() {
		if err := CheckTelegrafToml("cpu", []byte(s)); err != nil {
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
