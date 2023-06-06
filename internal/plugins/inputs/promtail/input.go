// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package promtail handles logs from promtail.
package promtail

import (
	"fmt"
	"net/http"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	inputName       = "promtail"
	measurementName = "default"
	catalogName     = "log"
	sampleConfig    = `
[inputs.promtail]
  #  以 legacy 版本接口处理请求时设置为 true, 对应 loki 的 API 为 /api/prom/push。
  legacy = false

  [inputs.promtail.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
)

type Input struct {
	Legacy bool              `toml:"legacy"`
	Tags   map[string]string `toml:"tags"`
	feeder dkio.Feeder
	opt    point.Option
}

type promtailSampleMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
	ipt    *Input
}

var l = logger.DefaultSLogger(inputName)

// Point implement MeasurementV2.
func (m *promtailSampleMeasurement) Point() *point.Point {
	opts := point.DefaultLoggingOptions()
	opts = append(opts, point.WithTime(m.ts), m.ipt.opt)

	return point.NewPointV2([]byte(m.name),
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (*promtailSampleMeasurement) LineProto() (*dkpt.Point, error) {
	// return nil, nil
	return nil, fmt.Errorf("not implement")
}

//nolint:lll
func (*promtailSampleMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "default",
		Type: "logging",
		Desc: "Using `source` field in the config file, default is `default`.",
		Tags: map[string]interface{}{
			"filename": inputs.NewTagInfo("File name. Optional."),
			"job":      inputs.NewTagInfo("Job name. Optional."),
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Message text, existed when default. Could use Pipeline to delete this field."}, // message
			"status":  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Log status."},
		},
	}
}

func (ipt *Input) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var (
		pipelinePath = getPipelinePath(req)
		source       = getSource(req)
		pts          []*point.Point
	)
	l.Debugf("receive log from %s, source = %s, pipeline = %s", req.URL.String(), source, pipelinePath)
	request, err := ipt.parseRequest(req)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		l.Errorf("failed to parse promtail request: %v", err)
		return
	}
	if len(request.Streams) == 0 {
		l.Warnf("receive empty request streams")
		resp.WriteHeader(http.StatusNoContent)
		return
	}
	now := time.Now()
	for _, s := range request.Streams {
		lbs, err := parseLabels(s.Labels)
		if err != nil {
			l.Warnf("failed to parse promtail labels: %v", err)
			continue
		}
		l.Debugf("got request stream with label string: %s, # parsed labels = %d", s.Labels, len(lbs))
		tags := make(map[string]string)
		for _, lb := range lbs {
			tags[lb.Name] = lb.Value
		}
		customTags := getCustomTags(req)
		for k, v := range customTags {
			tags[k] = v
		}
		for k, v := range ipt.Tags {
			tags[k] = v
		}

		for _, e := range s.Entries {
			logging := &promtailSampleMeasurement{
				name: source,
				tags: tags,
				fields: map[string]interface{}{
					pipeline.FieldMessage: e.Line,
					pipeline.FieldStatus:  pipeline.DefaultStatus,
				},
				ts:  now,
				ipt: ipt,
			}

			pts = append(pts, logging.Point())
		}
	}
	l.Debugf("received %d logs from promtail, feeding to io...", len(pts))

	if err := ipt.feeder.Feed(source, point.Logging, pts, &dkio.Option{PlScript: map[string]string{source: pipelinePath}}); err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
	} else {
		resp.WriteHeader(http.StatusNoContent)
	}
}

func (*Input) Catalog() string {
	return catalogName
}

func (ipt *Input) Run() {
	l.Info("register promtail router")

	// no election.
	ipt.opt = point.WithExtraTags(dkpt.GlobalHostTags())
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&promtailSampleMeasurement{}}
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (*Input) Terminate() {
	// Nothing to terminate.
}

func (ipt *Input) RegHTTPHandler() {
	l = logger.SLogger(inputName)
	httpapi.RegHTTPHandler("POST", "/v1/write/promtail", httpapi.ProtectedHandlerFunc(ipt.ServeHTTP, l))
}

func defaultInput() *Input {
	return &Input{
		Legacy: false,
		Tags:   map[string]string{},
		feeder: dkio.DefaultFeeder(),
	}
}

//nolint:gochecknoinits
func init() {
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
