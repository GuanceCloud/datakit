package inputs

import (
	"fmt"
	"testing"

	"github.com/influxdata/toml"

	"github.com/influxdata/telegraf/plugins/inputs/cpu"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func TestTomlMd5(t *testing.T) {
	c := cpu.CPUStats{
		PerCPU:         true,
		TotalCPU:       true,
		CollectCPUTime: true,
		ReportActive:   true,
	}

	x, err := datakit.TomlMd5(c)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(x)

	tomlStr := `
  ## Whether to report per-cpu stats or not
  percpu = true
  ## Whether to report total system cpu stats or not
  totalcpu = true
  ## If true, compute and report the sum of all non-idle CPU states.
  report_active = true

  ## If true, collect raw CPU time metrics.
  collect_cpu_time = true
	`

	tbl, err := toml.Parse([]byte(tomlStr))
	if err != nil {
		t.Fatal(err)
	}

	toml.UnmarshalTable(tbl, &c)

	x, err = datakit.TomlMd5(c)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(x)
}

func TestTelInput(t *testing.T) {
	cfg := `[[inputs.cpu]]

	## Whether to report per-cpu stats or not
	percpu = true
	## Whether to report total system cpu stats or not
	totalcpu = true
	## If true, collect raw CPU time metrics.
	collect_cpu_time = false
	## If true, compute and report the sum of all non-idle CPU states.
	report_active = false`

	result, err := TestTelegrafInput([]byte(cfg))
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("%s", string(result.Result))
}
