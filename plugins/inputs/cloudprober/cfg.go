package cloudprober

import (
	"github.com/influxdata/telegraf/plugins/common/tls"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"net/http"
)

type Input struct {
	URL string `toml:"url"`

	Interval datakit.Duration `toml:"interval"`

	tls.ClientConfig

	Tags map[string]string `toml:"tags"`

	client *http.Client
}
