// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package cloudprober scrape Google cloudprober metrics.
package cloudprober

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/prometheus/common/expfmt"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (*Input) SampleConfig() string {
	return sample
}

func (*Input) Catalog() string {
	return inputName
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Info("cloudprober start")
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
	client, err := ipt.createHTTPClient()
	if err != nil {
		l.Errorf("[error] cloudprober init client err:%s", err.Error())
		return
	}
	ipt.client = client

	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	for {
		ipt.getMetric()
		if ipt.lastErr != nil {
			dkio.FeedLastError(inputName, ipt.lastErr.Error(), point.Metric)
		}

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Info("cloudprober exit")
			return

		case <-ipt.semStop.Wait():
			l.Info("cloudprober return")
			return
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) getMetric() {
	resp, err := ipt.client.Get(ipt.URL)
	if err != nil {
		l.Errorf("error making HTTP request to %s: %s", ipt.URL, err)
		ipt.lastErr = err
		return
	}
	defer resp.Body.Close() //nolint:errcheck

	pts, err := ipt.parse(resp.Body)
	if err != nil {
		ipt.lastErr = err
		l.Error(err.Error())
		return
	}

	err = ipt.feeder.Feed(inputName, point.Metric, pts,
		&dkio.Option{CollectCost: time.Since(ipt.start)})
	if err != nil {
		l.Error(err.Error())
		ipt.lastErr = err
	}
}

func (ipt *Input) parse(reader io.Reader) ([]*point.Point, error) {
	var (
		parse expfmt.TextParser
		pts   []*point.Point
	)
	Family, err := parse.TextToMetricFamilies(reader)
	if err != nil {
		return pts, err
	}
	for metricName, family := range Family {
		for _, metric := range family.Metric {
			measurement := &Measurement{
				tags:   map[string]string{},
				fields: map[string]interface{}{},
				ts:     datakit.TimestampMsToTime(metric.GetTimestampMs()),
			}
			for k, v := range ipt.Tags {
				measurement.tags[k] = v
			}
			for _, label := range metric.Label {
				if label.GetName() == "ptype" {
					measurement.name = fmt.Sprintf("probe_%s", label.GetValue())
					continue
				}
				measurement.tags[label.GetName()] = label.GetValue()
			}
			switch family.GetType().String() {
			case "COUNTER":
				measurement.fields[metricName] = metric.Counter.GetValue()
			case "GAUGE":
				measurement.fields[metricName] = metric.Gauge.GetValue()
			case "SUMMARY":
				measurement.fields[metricName] = metric.Summary.GetSampleCount()
			case "UNTYPED":
				measurement.fields[metricName] = metric.Untyped.GetValue()
			case "HISTOGRAM":
				measurement.fields[metricName] = metric.Histogram.GetSampleCount()
			}

			opts := point.DefaultMetricOptions()
			opts = append(opts, point.WithTime(measurement.ts))

			measurement.tags = inputs.MergeTags(ipt.Tagger.HostTags(), measurement.tags, ipt.URL)

			pt := point.NewPointV2(measurement.name,
				append(point.NewTags(measurement.tags), point.NewKVs(measurement.fields)...),
				opts...)
			pts = append(pts, pt)
		}
	}
	return pts, nil
}

func (ipt *Input) createHTTPClient() (*http.Client, error) {
	tlsCfg, err := ipt.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
	}

	return client, nil
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&Measurement{},
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		s := &Input{
			Interval: datakit.Duration{Duration: time.Second * 5},

			feeder:  dkio.DefaultFeeder(),
			semStop: cliutils.NewSem(),
			Tagger:  datakit.DefaultGlobalTagger(),
		}
		return s
	})
}
