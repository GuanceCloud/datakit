package jenkins

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	"github.com/influxdata/telegraf/plugins/common/tls"

	"net/http"
	"sync"
	"time"
)

type Input struct {
	Url             string               `toml:"url"`
	Key             string               `toml:"key"`
	Interval        datakit.Duration     `toml:"interval"`
	ResponseTimeout datakit.Duration     `toml:"response_timeout"`
	Log             *inputs.TailerOption `toml:"log"`
	Tags            map[string]string    `toml:"tags"`

	tls.ClientConfig
	// HTTP client
	client *http.Client

	start time.Time
	tail  *inputs.Tailer

	lastErr      error
	wg           sync.WaitGroup
	collectCache []inputs.Measurement
}

func newCountFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}
