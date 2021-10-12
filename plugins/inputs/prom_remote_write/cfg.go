package prom_remote_write

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "prom_remote_write"
	catalog   = "prom"
	sample    = `
[[inputs.prom_remote_write]]
  ## Path to listen to.
  path = "/receive"

  ## accepted methods
  methods = ["PUT", "POST"]

  ## Part of the request to consume.  Available options are "body" and
  ## "query".
  # data_source = "body"

  ## metric name filter
  # regex is supported
  # no filter if empty
  # metric_name_filter = ["gc", "go"]

  ## metric name prefix
  # prefix will be added to metric name
  # measurement_prefix = "prefix_"

  ## metric name
  # metric name will be divided by "_" by default.
  # metric is named by the first divided field, the remaining field is used as the current metric name
  # metric name will not be divided if measurement_name is configured
  # measurement_prefix will be added to the start of measurement_name
  # measurement_name = "prom"

  ## tags to ignore
  # tags_ignore = ["xxxx"]

  ## max body size in bytes, default set to 500MB
  # max_body_size = 0

  ## Optional username and password to accept for HTTP basic authentication.
  ## You probably want to make sure you have TLS configured above for this.
  # basic_username = ""
  # basic_password = ""

  ## Optional setting to map http headers into tags
  ## If the http header is not present on the request, no corresponding tag will be added
  ## If multiple instances of the http header are present, only the first value will be used
  [inputs.prom_remote_write.http_header_tags]
  # HTTP_HEADER = "TAG_NAME"

  ## 自定义Tags
  [inputs.prom_remote_write.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
)

// defaultMaxBodySize is the default maximum request body size, in bytes.
// if the request body is over this size, we will return an HTTP 413 error.
// 500 MB
const defaultMaxBodySize int64 = 500 * 1024 * 1024

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
		Name:   inputName,
		Desc:   "prometheus remote write指标",
		Fields: map[string]interface{}{},
		Tags:   map[string]interface{}{},
	}
}
