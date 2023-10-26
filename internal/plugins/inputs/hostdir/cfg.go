// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostdir

import (
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
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
	Dir             string           `toml:"dir"`
	ExcludePatterns []string         `toml:"exclude_patterns"`
	Interval        datakit.Duration `toml:"interval"`
	collectCache    []*point.Point
	Tags            map[string]string `toml:"tags"`
	platform        string

	semStop *cliutils.Sem // start stop signal
	feeder  dkio.Feeder
	Tagger  datakit.GlobalTagger
}

type Measurement struct{}

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
