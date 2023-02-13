// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package promtail handles logs from promtail.
package promtail

import (
	"net/http"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	ihttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/http"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName    = "promtail"
	catalogName  = "log"
	sampleConfig = `
[inputs.promtail]
  #  以 legacy 版本接口处理请求时设置为 true，对应 loki 的 API 为 /api/prom/push。
  legacy = false

  [inputs.promtail.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
)

type Input struct {
	Legacy bool              `toml:"legacy"`
	Tags   map[string]string `toml:"tags"`
}

type promtailSampleMeasurement struct{}

var l = logger.DefaultSLogger(inputName)

func (p *promtailSampleMeasurement) LineProto() (*point.Point, error) {
	return nil, nil
}

func (p *promtailSampleMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "promtail 日志接收",
		Desc:   "",
		Type:   "logging",
		Fields: map[string]interface{}{},
		Tags:   map[string]interface{}{},
	}
}

func (i *Input) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var (
		pipelinePath = getPipelinePath(req)
		source       = getSource(req)
		pts          []*point.Point
	)
	l.Debugf("receive log from %s, source = %s, pipeline = %s", req.URL.String(), source, pipelinePath)
	request, err := i.parseRequest(req)
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
		for k, v := range i.Tags {
			tags[k] = v
		}

		for _, e := range s.Entries {
			pt, err := point.NewPoint(source, tags, map[string]interface{}{
				pipeline.FieldMessage: e.Line,
				pipeline.FieldStatus:  pipeline.DefaultStatus,
			}, point.LOpt())
			if err != nil {
				l.Error(err)
			} else {
				pts = append(pts, pt)
			}
		}
	}
	l.Debugf("received %d logs from promtail, feeding to io...", len(pts))
	if err := dkio.Feed(source, datakit.Logging, pts, &dkio.Option{PlScript: map[string]string{source: pipelinePath}}); err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
	} else {
		resp.WriteHeader(http.StatusNoContent)
	}
}

func (i *Input) Catalog() string {
	return catalogName
}

func (i *Input) Run() {
	l.Info("register promtail router")
}

func (i *Input) SampleConfig() string {
	return sampleConfig
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&promtailSampleMeasurement{}}
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (i *Input) Terminate() {
	// Nothing to terminate.
}

func (i *Input) RegHTTPHandler() {
	l = logger.SLogger(inputName)
	dhttp.RegHTTPHandler("POST", "/v1/write/promtail", ihttp.ProtectedHandlerFunc(i.ServeHTTP, l))
}

//nolint:gochecknoinits
func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Legacy: false,
			Tags:   map[string]string{},
		}
	})
}
