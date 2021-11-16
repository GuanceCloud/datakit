package hostdir

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName   = "hostdir"
	metricName  = inputName
	l           = logger.DefaultSLogger(inputName)
	minInterval = time.Second
	maxInterval = time.Second * 30
	sample      = `
[[inputs.hostdir]]
  interval = "10s"

  # directory to collect
  # Windows example: C:\\Users
  # UNIX-like example: /usr/local/
  dir = "" # required

	# optional, i.e., "*.exe", "*.so"
  exclude_patterns = []

[inputs.hostdir.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"`
)

type Input struct {
	Dir string `toml:"dir"`
	// file_size string
	// file_count      string
	ExcludePatterns []string         `toml:"exclude_patterns"`
	Interval        datakit.Duration `toml:"interval"`
	collectCache    []inputs.Measurement
	Tags            map[string]string `toml:"tags"`
	platform        string

	semStop *cliutils.Sem // start stop signal
}

type Measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *Measurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Fields: map[string]interface{}{
			"file_size":  newCountFieldInfo("The size of files"),
			"file_count": newCountFieldInfo("The number of files"),
			"dir_count":  newCountFieldInfo("The number of Dir"),
		},
		Tags: map[string]interface{}{
			"host_directory": inputs.NewTagInfo("the start Dir"),
			"file_ownership": inputs.NewTagInfo("file ownership"),
			"file_system":    inputs.NewTagInfo("file system type"),
		},
	}
}

func newCountFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}
