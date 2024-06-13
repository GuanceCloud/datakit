// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package pushgateway wraps the Pushgateway API.
package pushgateway

import (
	"errors"
	"io"
	"net/http"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	dto "github.com/prometheus/client_model/go"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	iprom "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
)

var (
	_ inputs.HTTPInput = (*Input)(nil)
	_ inputs.ReadEnv   = (*Input)(nil)
	_ inputs.Singleton = (*Input)(nil)
)

const (
	inputName = "pushgateway"

	sampleConfig = `
[[inputs.pushgateway]]
  ## Prefix for the internal routes of web endpoints. Defaults to empty.
  route_prefix = ""

  ## Measurement name.
  ## If measurement_name is empty, split metric name by '_', the first field after split as measurement set name, the rest as current metric name.
  ## If measurement_name is not empty, using this as measurement set name.
  # measurement_name = "prom_pushgateway"

  ## Keep Exist Metric Name.
  ## If the keep_exist_metric_name is true, keep the raw value for field names.
  keep_exist_metric_name = true
`
)

var log = logger.DefaultSLogger(inputName)

type Input struct {
	RoutePrefix         string `toml:"route_prefix,omitempty"`
	MeasurementName     string `toml:"measurement_name"`
	KeepExistMetricName bool   `toml:"keep_exist_metric_name"`
	feeder              dkio.Feeder
}

func (*Input) SampleConfig() string                    { return sampleConfig }
func (*Input) Catalog() string                         { return inputName }
func (*Input) AvailableArchs() []string                { return datakit.AllOS }
func (*Input) Singleton()                              { /*nil*/ }
func (*Input) Run()                                    { /*nil*/ }
func (*Input) SampleMeasurement() []inputs.Measurement { return nil /* no measurement docs exported */ }
func (*Input) Terminate()                              { /* TODO */ }

func (ipt *Input) RegHTTPHandler() {
	log = logger.SLogger(inputName)

	opts := []iprom.PromOption{
		iprom.WithLogger(log), // WithLogger must in the first
		iprom.WithSource(inputName),
		iprom.KeepExistMetricName(ipt.KeepExistMetricName),
	}
	if ipt.MeasurementName != "" {
		opts = append(opts, iprom.WithMeasurementName(ipt.MeasurementName))
	}

	text := func(body io.Reader, tags map[string]string) error {
		return textProcessor(opts, ipt.feeder, body, tags)
	}
	protobuf := func(body io.Reader, tags map[string]string) error {
		return protobufProcessor(opts, ipt.feeder, body, tags)
	}

	path := ipt.RoutePrefix + "/metrics"

	for _, suffix := range []string{"", base64Suffix} {
		jobBase64Encoded := suffix == base64Suffix
		httpapi.RegHTTPRoute(http.MethodPost, path+"/job"+suffix+"/:job", pushHandle(jobBase64Encoded, text, protobuf))
		httpapi.RegHTTPRoute(http.MethodPut, path+"/job"+suffix+"/:job", pushHandle(jobBase64Encoded, text, protobuf))
		httpapi.RegHTTPRoute(http.MethodPost, path+"/job"+suffix+"/:job/*labels", pushHandle(jobBase64Encoded, text, protobuf))
		httpapi.RegHTTPRoute(http.MethodPut, path+"/job"+suffix+"/:job/*labels", pushHandle(jobBase64Encoded, text, protobuf))
	}
}

func textProcessor(opts []iprom.PromOption, feeder dkio.Feeder, body io.Reader, tags map[string]string) error {
	pm, err := iprom.NewProm(opts...)
	if err != nil {
		log.Errorf("new prom failed: %s", err)
		return err
	}

	pts, err := pm.ProcessMetrics(body, "")
	if err != nil {
		return err
	}

	addTagsToPoints(pts, tags)
	return feeder.FeedV2(point.Metric, pts, dkio.WithInputName(inputName), dkio.DisableGlobalTags(true))
}

func protobufProcessor(opts []iprom.PromOption, feeder dkio.Feeder, body io.Reader, tags map[string]string) error {
	var err error

	metricFamilies := map[string]*dto.MetricFamily{}
	for {
		mf := &dto.MetricFamily{}
		if _, err = pbutil.ReadDelimited(body, mf); err != nil {
			if errors.Is(err, io.EOF) {
				//nolint:ineffassign
				err = nil
			}
			break
		}
		metricFamilies[mf.GetName()] = mf
	}

	pm, err := iprom.NewProm(opts...)
	if err != nil {
		log.Errorf("new prom failed: %s", err)
		return err
	}

	pts, err := pm.MetricFamilies2points(metricFamilies, "")
	if err != nil {
		return err
	}

	addTagsToPoints(pts, tags)
	return feeder.FeedV2(point.Metric, pts, dkio.WithInputName(inputName), dkio.DisableGlobalTags(true))
}

func addTagsToPoints(pts []*point.Point, tags map[string]string) {
	for _, pt := range pts {
		pt.AddKVs(point.NewTags(tags)...)
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			feeder: dkio.DefaultFeeder(),
		}
	})
}
